package tools

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/flowfiles"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

// contextTestTermLogProvider implements TermLogProvider for context tests.
type contextTestTermLogProvider struct{}

func (m *contextTestTermLogProvider) PutMsg(_ context.Context, _ database.TermlogType, _ string,
	_ int64, _, _ *int64) (int64, error) {
	return 1, nil
}

var _ TermLogProvider = (*contextTestTermLogProvider)(nil)

// contextAwareMockDockerClient tracks whether the context was canceled
// when getExecResult runs, proving context.WithoutCancel works.
type contextAwareMockDockerClient struct {
	isRunning      bool
	execCreateResp container.ExecCreateResponse
	attachOutput   []byte
	attachDelay    time.Duration
	inspectResp    container.ExecInspect

	// Set by ContainerExecAttach to track if ctx was canceled during attach
	ctxWasCanceled bool
}

func (m *contextAwareMockDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (m *contextAwareMockDockerClient) StopContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *contextAwareMockDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *contextAwareMockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return m.isRunning, nil
}
func (m *contextAwareMockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return m.execCreateResp, nil
}
func (m *contextAwareMockDockerClient) ContainerExecAttach(ctx context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	// Wait for the configured delay, simulating a long-running command
	if m.attachDelay > 0 {
		select {
		case <-time.After(m.attachDelay):
			// Command completed normally
		case <-ctx.Done():
			// Context was canceled -- this is the bug behavior (without WithoutCancel)
			m.ctxWasCanceled = true
			return types.HijackedResponse{}, ctx.Err()
		}
	}

	// Check if context was already canceled by the time we get here
	select {
	case <-ctx.Done():
		m.ctxWasCanceled = true
		return types.HijackedResponse{}, ctx.Err()
	default:
	}

	pr, pw := net.Pipe()
	go func() {
		pw.Write(m.attachOutput)
		pw.Close()
	}()

	return types.HijackedResponse{
		Conn:   pr,
		Reader: bufio.NewReader(pr),
	}, nil
}
func (m *contextAwareMockDockerClient) ContainerStatPath(_ context.Context, _ string, _ string) (container.PathStat, error) {
	return container.PathStat{}, nil
}
func (m *contextAwareMockDockerClient) ListContainerDir(_ context.Context, _ string, _ string) ([]container.PathStat, error) {
	return nil, nil
}
func (m *contextAwareMockDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return m.inspectResp, nil
}
func (m *contextAwareMockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, _ io.Reader, _ container.CopyToContainerOptions) error {
	return nil
}
func (m *contextAwareMockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(nil), container.PathStat{}, nil
}
func (m *contextAwareMockDockerClient) Cleanup(_ context.Context) error { return nil }
func (m *contextAwareMockDockerClient) GetDefaultImage() string         { return "test-image" }

var _ docker.DockerClient = (*contextAwareMockDockerClient)(nil)

func TestExecCommandDetachSurvivesParentCancel(t *testing.T) {
	// This test validates the fix for Issue #176:
	// Detached commands must NOT be killed when the parent context is canceled.
	//
	// Before the fix: detached goroutine used parent ctx directly, so when the
	// parent was canceled (e.g., agent delegation timeout), ctx.Done() fired
	// in getExecResult and killed the background command.
	//
	// After the fix: context.WithoutCancel(ctx) creates an isolated context
	// that preserves values but ignores parent cancellation.

	mock := &contextAwareMockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-cancel-test"},
		attachOutput:   []byte("background result"),
		attachDelay:    2 * time.Second, // simulates a long-running command
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	// Create a cancellable parent context
	parentCtx, cancel := context.WithCancel(t.Context())

	// Start ExecCommand with detach=true (returns quickly due to quick check timeout)
	output, err := term.ExecCommand(parentCtx, "/work", "long-running-scan", true, 5*time.Minute)
	assert.NoError(t, err)
	assert.Contains(t, output, "Command started in background")

	// Cancel the parent context -- simulating agent delegation timeout
	cancel()

	// Wait enough time for the detached goroutine to complete its work.
	// If context.WithoutCancel is working correctly, the goroutine should
	// NOT see ctx.Done() and should complete normally after attachDelay.
	// If the fix regresses, ctxWasCanceled will be true.
	time.Sleep(3 * time.Second)

	assert.False(t, mock.ctxWasCanceled,
		"detached goroutine should NOT see parent context cancellation (context.WithoutCancel must be used)")
}

func TestExecCommandNonDetachRespectsParentCancel(t *testing.T) {
	// Counterpart: non-detached commands SHOULD respect parent cancellation.
	// This ensures we didn't accidentally apply WithoutCancel to the non-detach path.

	mock := &contextAwareMockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-nondetach-cancel"},
		attachOutput:   []byte("should not complete"),
		attachDelay:    5 * time.Second, // longer than cancel delay
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	parentCtx, cancel := context.WithCancel(t.Context())

	// Cancel after 200ms -- non-detached command should see this
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_, err := term.ExecCommand(parentCtx, "/work", "long-command", false, 5*time.Minute)

	// Non-detached command should fail with context error
	assert.Error(t, err)
	assert.True(t, mock.ctxWasCanceled,
		"non-detached command SHOULD see parent context cancellation")
}

func TestPrimaryTerminalName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		flowID int64
		want   string
	}{
		{1, PrimaryTerminalNamePrefix + "1"},
		{0, PrimaryTerminalNamePrefix + "0"},
		{12345, PrimaryTerminalNamePrefix + "12345"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("flowID=%d", tt.flowID), func(t *testing.T) {
			t.Parallel()

			if got := PrimaryTerminalName(tt.flowID); got != tt.want {
				t.Errorf("PrimaryTerminalName(%d) = %q, want %q", tt.flowID, got, tt.want)
			}
		})
	}
}

func TestWriteUploadsTar(t *testing.T) {
	uploadDir := t.TempDir()
	requireNoError := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	requireNoError(os.WriteFile(filepath.Join(uploadDir, "a.txt"), []byte("alpha"), 0644))
	requireNoError(os.Mkdir(filepath.Join(uploadDir, "sub"), 0755))
	requireNoError(os.WriteFile(filepath.Join(uploadDir, "sub", "b.txt"), []byte("bravo"), 0644))
	if err := os.Symlink(filepath.Join(uploadDir, "a.txt"), filepath.Join(uploadDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- flowfiles.WriteUploadsTar(pw, uploadDir)
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, pr)
	assert.NoError(t, err)
	assert.NoError(t, <-errCh)

	tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
	contents := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		assert.NoError(t, err)
		contents[hdr.Name] = string(data)
	}

	assert.Equal(t, "alpha", contents["uploads/a.txt"])
	assert.Equal(t, "bravo", contents["uploads/sub/b.txt"])
	assert.NotContains(t, contents, "uploads/link.txt")
}

func TestCollectFileSyncEntries(t *testing.T) {
	localDir := t.TempDir()
	requireNoError := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	requireNoError(os.MkdirAll(filepath.Join(localDir, "targets"), 0755))
	requireNoError(os.WriteFile(filepath.Join(localDir, "targets", "ips.txt"), []byte("127.0.0.1"), 0644))
	requireNoError(os.WriteFile(filepath.Join(localDir, "top.txt"), []byte("top"), 0644))
	if err := os.Symlink(filepath.Join(localDir, "top.txt"), filepath.Join(localDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	entries, err := collectFileSyncEntries(localDir, flowfiles.UploadsDirName)
	assert.NoError(t, err)

	byTarPath := map[string]fileSyncEntry{}
	for _, entry := range entries {
		byTarPath[entry.tarPath] = entry
	}

	assert.Contains(t, byTarPath, "uploads/top.txt")
	assert.Contains(t, byTarPath, "uploads/targets/ips.txt")
	assert.NotContains(t, byTarPath, "uploads/link.txt")
	assert.Equal(t, docker.WorkFolderPathInContainer+"/uploads/top.txt", byTarPath["uploads/top.txt"].containerPath)
	assert.Equal(t, filepath.Join(localDir, "top.txt"), byTarPath["uploads/top.txt"].localPath)
}

func TestCollectFileSyncEntriesMissingDir(t *testing.T) {
	entries, err := collectFileSyncEntries(filepath.Join(t.TempDir(), "missing"), flowfiles.ResourcesDirName)
	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestConvertSyncEntriesToTarEntries(t *testing.T) {
	entries := []fileSyncEntry{
		{localPath: "/tmp/a.txt", tarPath: "uploads/a.txt"},
		{localPath: "/tmp/b.txt", tarPath: "resources/b.txt"},
	}

	tarEntries := convertSyncEntriesToTarEntries(entries)

	assert.Equal(t, []flowfiles.TarEntry{
		{LocalPath: "/tmp/a.txt", TarPath: "uploads/a.txt"},
		{LocalPath: "/tmp/b.txt", TarPath: "resources/b.txt"},
	}, tarEntries)
}

func TestFindMissingInContainer(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		execCreateResp: container.ExecCreateResponse{ID: "exec-file-check"},
		attachOutput: []byte(
			docker.WorkFolderPathInContainer + "/uploads/a.txt\n" +
				docker.WorkFolderPathInContainer + "/resources/b.txt\n" +
				docker.WorkFolderPathInContainer + "/unknown.txt\n",
		),
		inspectResp: container.ExecInspect{ExitCode: 0},
	}
	fte := &flowToolsExecutor{flowID: 7, docker: mock}
	entries := []fileSyncEntry{
		{localPath: "/tmp/a.txt", containerPath: docker.WorkFolderPathInContainer + "/uploads/a.txt", tarPath: "uploads/a.txt"},
		{localPath: "/tmp/b.txt", containerPath: docker.WorkFolderPathInContainer + "/resources/b.txt", tarPath: "resources/b.txt"},
	}

	missing, err := fte.findMissingInContainer(t.Context(), entries)

	assert.NoError(t, err)
	assert.Equal(t, entries, missing)
}

func TestFindMissingInContainerChecksExitCode(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		execCreateResp: container.ExecCreateResponse{ID: "exec-file-check"},
		attachOutput:   []byte("shell failed"),
		inspectResp:    container.ExecInspect{ExitCode: 2},
	}
	fte := &flowToolsExecutor{flowID: 7, docker: mock}

	_, err := fte.findMissingInContainer(t.Context(), []fileSyncEntry{
		{containerPath: docker.WorkFolderPathInContainer + "/uploads/a.txt"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit code 2")
}

func TestConfiguredExecTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured time.Duration
		want       time.Duration
	}{
		{
			name:       "typical value is returned as-is",
			configured: 600 * time.Second,
			want:       600 * time.Second,
		},
		{
			name:       "new default (1200 s) is returned as-is",
			configured: 1200 * time.Second,
			want:       1200 * time.Second,
		},
		{
			name:       "exactly at the 3-hour ceiling is returned as-is",
			configured: maxExplicitExecCommandTimeout,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "zero is capped to the 3-hour ceiling",
			configured: 0,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "negative one second is capped to the 3-hour ceiling",
			configured: -1 * time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "large negative is capped to the 3-hour ceiling",
			configured: -9999 * time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "one second above the ceiling is capped",
			configured: maxExplicitExecCommandTimeout + time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "very large value (> 3 h) is capped to the 3-hour ceiling",
			configured: 100000 * time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			term := &terminal{defaultExecTimeout: tt.configured}
			assert.Equal(t, tt.want, term.configuredExecTimeout())
		})
	}
}

func TestNormalizeExecTimeout(t *testing.T) {
	t.Parallel()

	// ceilFor computes the effective runtime ceiling for a given configured value:
	// it equals configuredExecTimeout() + defaultExtraExecTimeout.
	ceilFor := func(configured time.Duration) time.Duration {
		term := &terminal{defaultExecTimeout: configured}
		return term.configuredExecTimeout() + defaultExtraExecTimeout
	}

	tests := []struct {
		name       string
		configured time.Duration
		requested  time.Duration
		want       time.Duration
	}{
		// --- Explicit positive values: preserved when within the operator ceiling ---
		{
			name:       "typical explicit value is preserved",
			configured: 10 * time.Minute,
			requested:  45 * time.Second,
			want:       45 * time.Second,
		},
		{
			name:       "explicit value exactly at the ceiling is preserved",
			configured: 10 * time.Minute,
			requested:  ceilFor(10 * time.Minute), // 600s + 5s = 605s
			want:       ceilFor(10 * time.Minute),
		},
		{
			name:       "explicit value one second above the ceiling falls back to ceiling",
			configured: 10 * time.Minute,
			requested:  ceilFor(10*time.Minute) + time.Second,
			want:       ceilFor(10 * time.Minute),
		},
		{
			name:       "explicit value at the default configured (1200 s) is preserved",
			configured: 1200 * time.Second,
			requested:  1200 * time.Second,
			want:       1200 * time.Second,
		},
		{
			name:       "explicit value above the 1200-s ceiling falls back to that ceiling",
			configured: 1200 * time.Second,
			requested:  ceilFor(1200*time.Second) + time.Second, // 1205s + 1s → fallback
			want:       ceilFor(1200 * time.Second),             // 1205s
		},
		{
			name:       "explicit value at the 3-hour ceiling is preserved when configured=0",
			configured: 0,
			requested:  ceilFor(0), // 3h + 5s
			want:       ceilFor(0),
		},
		{
			name:       "explicit value above the 3-hour ceiling falls back to 3-hour ceiling",
			configured: 0,
			requested:  ceilFor(0) + time.Second,
			want:       ceilFor(0),
		},

		// --- Zero requested: falls back to the operator ceiling ---
		{
			name:       "zero requested with typical configured falls back to ceiling",
			configured: 10 * time.Minute,
			requested:  0,
			want:       ceilFor(10 * time.Minute), // 605s
		},
		{
			name:       "zero requested with default configured (1200 s) falls back to ceiling",
			configured: 1200 * time.Second,
			requested:  0,
			want:       ceilFor(1200 * time.Second), // 1205s
		},
		{
			name:       "zero requested with configured=0 falls back to 3-hour ceiling",
			configured: 0,
			requested:  0,
			want:       ceilFor(0), // 3h + 5s
		},
		{
			name:       "zero requested with oversized configured (> 3 h) falls back to 3-hour ceiling",
			configured: 100000 * time.Second,
			requested:  0,
			want:       ceilFor(0), // capped to 3h + 5s
		},

		// --- Negative requested: treated identically to zero ---
		{
			name:       "negative requested falls back to configured ceiling",
			configured: 10 * time.Minute,
			requested:  -5 * time.Second,
			want:       ceilFor(10 * time.Minute),
		},
		{
			name:       "negative requested with configured=0 falls back to 3-hour ceiling",
			configured: 0,
			requested:  -1 * time.Second,
			want:       ceilFor(0),
		},
		{
			name:       "negative requested with negative configured falls back to 3-hour ceiling",
			configured: -5 * time.Second,
			requested:  -1 * time.Second,
			want:       ceilFor(0), // both negative → absolute 3-hour max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			term := &terminal{defaultExecTimeout: tt.configured}
			assert.Equal(t, tt.want, term.normalizeExecTimeout(tt.requested))
		})
	}
}

// TestHandleDoesNotDoublePadTimeout verifies that Handle() does not add
// defaultExtraExecTimeout on top of what normalizeExecTimeout already returns.
// This test validates the fix for the double timeout padding bug.
func TestHandleDoesNotDoublePadTimeout(t *testing.T) {
	// Create a mock that tracks the timeout passed to ExecCommand
	var capturedTimeout time.Duration
	mock := &timeoutCapturingMockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-timeout-test"},
		attachOutput:   []byte("output"),
		inspectResp:    container.ExecInspect{ExitCode: 0},
		onExecCreate: func(timeout time.Duration) {
			capturedTimeout = timeout
		},
	}

	term := &terminal{
		flowID:             1,
		containerID:        1,
		containerLID:       "test-container",
		dockerClient:       mock,
		tlp:                &contextTestTermLogProvider{},
		defaultExecTimeout: 60 * time.Second, // configured timeout
	}

	// Request a 10 second timeout
	action := TerminalAction{
		Input:   "echo test",
		Timeout: 10,
	}
	args, _ := json.Marshal(action)

	_, err := term.Handle(t.Context(), TerminalToolName, args)
	assert.NoError(t, err)

	// The timeout passed to ExecCommand should be normalized (10s),
	// NOT normalized + 5s (15s). ExecCommand will normalize again,
	// so if Handle() added extra padding, the final timeout would be wrong.
	// With configured=60s, normalizeExecTimeout(10s) returns 10s (since 10s <= 65s ceiling).
	// The bug was adding another 5s here, making it 15s.
	assert.Equal(t, 10*time.Second, capturedTimeout,
		"Handle() should not add defaultExtraExecTimeout on top of normalized timeout")
}

// TestWriteFileLogsSuccessAsStdout verifies that WriteFile() logs the success
// message as TermlogTypeStdout, not TermlogTypeStdin.
func TestWriteFileLogsSuccessAsStdout(t *testing.T) {
	var capturedType database.TermlogType
	mock := &logTypeCapturingMockDockerClient{
		isRunning: true,
		onPutMsg: func(msgType database.TermlogType) {
			capturedType = msgType
		},
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &logTypeCapturingTermLogProvider{onPutMsg: func(msgType database.TermlogType) { capturedType = msgType }},
	}

	_, err := term.WriteFile(t.Context(), 1, "test content", "/tmp/test.txt")
	assert.NoError(t, err)

	// The success message should be logged as stdout, not stdin
	assert.Equal(t, database.TermlogTypeStdout, capturedType,
		"WriteFile() should log success message as TermlogTypeStdout, not TermlogTypeStdin")
}

// TestReadFileUsesReadFull verifies that ReadFile() uses io.ReadFull to ensure
// all bytes are read, not just a single Read() call that might return fewer bytes.
func TestReadFileUsesReadFull(t *testing.T) {
	// Create a tar archive with a file
	var tarBuffer bytes.Buffer
	tarWriter := tar.NewWriter(&tarBuffer)

	content := "test file content that should be fully read"
	err := tarWriter.WriteHeader(&tar.Header{
		Name: "test.txt",
		Mode: 0600,
		Size: int64(len(content)),
	})
	assert.NoError(t, err)
	_, err = tarWriter.Write([]byte(content))
	assert.NoError(t, err)
	err = tarWriter.Close()
	assert.NoError(t, err)

	// Create a mock that returns the tar archive
	mock := &readFullVerifyingMockDockerClient{
		isRunning:   true,
		tarResponse: tarBuffer.Bytes(),
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	result, err := term.ReadFile(t.Context(), 1, "/tmp/test.txt")
	assert.NoError(t, err)
	assert.Contains(t, result, content, "ReadFile() should read all bytes of the file")
}

// Mock implementations for the new tests

type timeoutCapturingMockDockerClient struct {
	isRunning      bool
	execCreateResp container.ExecCreateResponse
	attachOutput   []byte
	inspectResp    container.ExecInspect
	onExecCreate   func(timeout time.Duration)
}

func (m *timeoutCapturingMockDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (m *timeoutCapturingMockDockerClient) StopContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *timeoutCapturingMockDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *timeoutCapturingMockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return m.isRunning, nil
}
func (m *timeoutCapturingMockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return m.execCreateResp, nil
}
func (m *timeoutCapturingMockDockerClient) ContainerExecAttach(_ context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	pr, pw := net.Pipe()
	go func() {
		pw.Write(m.attachOutput)
		pw.Close()
	}()
	return types.HijackedResponse{
		Conn:   pr,
		Reader: bufio.NewReader(pr),
	}, nil
}
func (m *timeoutCapturingMockDockerClient) ContainerStatPath(_ context.Context, _ string, _ string) (container.PathStat, error) {
	return container.PathStat{}, nil
}
func (m *timeoutCapturingMockDockerClient) ListContainerDir(_ context.Context, _ string, _ string) ([]container.PathStat, error) {
	return nil, nil
}
func (m *timeoutCapturingMockDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return m.inspectResp, nil
}
func (m *timeoutCapturingMockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, _ io.Reader, _ container.CopyToContainerOptions) error {
	return nil
}
func (m *timeoutCapturingMockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(nil), container.PathStat{}, nil
}
func (m *timeoutCapturingMockDockerClient) Cleanup(_ context.Context) error { return nil }
func (m *timeoutCapturingMockDockerClient) GetDefaultImage() string         { return "test-image" }

var _ docker.DockerClient = (*timeoutCapturingMockDockerClient)(nil)

type logTypeCapturingMockDockerClient struct {
	isRunning bool
	onPutMsg  func(msgType database.TermlogType)
}

func (m *logTypeCapturingMockDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (m *logTypeCapturingMockDockerClient) StopContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *logTypeCapturingMockDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *logTypeCapturingMockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return m.isRunning, nil
}
func (m *logTypeCapturingMockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{}, nil
}
func (m *logTypeCapturingMockDockerClient) ContainerExecAttach(_ context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, nil
}
func (m *logTypeCapturingMockDockerClient) ContainerStatPath(_ context.Context, _ string, _ string) (container.PathStat, error) {
	return container.PathStat{}, nil
}
func (m *logTypeCapturingMockDockerClient) ListContainerDir(_ context.Context, _ string, _ string) ([]container.PathStat, error) {
	return nil, nil
}
func (m *logTypeCapturingMockDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return container.ExecInspect{}, nil
}
func (m *logTypeCapturingMockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, _ io.Reader, _ container.CopyToContainerOptions) error {
	return nil
}
func (m *logTypeCapturingMockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(nil), container.PathStat{}, nil
}
func (m *logTypeCapturingMockDockerClient) Cleanup(_ context.Context) error { return nil }
func (m *logTypeCapturingMockDockerClient) GetDefaultImage() string         { return "test-image" }

var _ docker.DockerClient = (*logTypeCapturingMockDockerClient)(nil)

type logTypeCapturingTermLogProvider struct {
	onPutMsg func(msgType database.TermlogType)
}

func (m *logTypeCapturingTermLogProvider) PutMsg(_ context.Context, msgType database.TermlogType, _ string,
	_ int64, _, _ *int64) (int64, error) {
	if m.onPutMsg != nil {
		m.onPutMsg(msgType)
	}
	return 1, nil
}

var _ TermLogProvider = (*logTypeCapturingTermLogProvider)(nil)

type readFullVerifyingMockDockerClient struct {
	isRunning   bool
	tarResponse []byte
}

func (m *readFullVerifyingMockDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (m *readFullVerifyingMockDockerClient) StopContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *readFullVerifyingMockDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *readFullVerifyingMockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return m.isRunning, nil
}
func (m *readFullVerifyingMockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{}, nil
}
func (m *readFullVerifyingMockDockerClient) ContainerExecAttach(_ context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, nil
}
func (m *readFullVerifyingMockDockerClient) ContainerStatPath(_ context.Context, _ string, _ string) (container.PathStat, error) {
	return container.PathStat{Mode: 0}, nil
}
func (m *readFullVerifyingMockDockerClient) ListContainerDir(_ context.Context, _ string, _ string) ([]container.PathStat, error) {
	return nil, nil
}
func (m *readFullVerifyingMockDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return container.ExecInspect{}, nil
}
func (m *readFullVerifyingMockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, _ io.Reader, _ container.CopyToContainerOptions) error {
	return nil
}
func (m *readFullVerifyingMockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(bytes.NewReader(m.tarResponse)), container.PathStat{Mode: 0}, nil
}
func (m *readFullVerifyingMockDockerClient) Cleanup(_ context.Context) error { return nil }
func (m *readFullVerifyingMockDockerClient) GetDefaultImage() string         { return "test-image" }

var _ docker.DockerClient = (*readFullVerifyingMockDockerClient)(nil)
