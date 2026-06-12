#!/bin/bash
set -e

RESOLUTION="${VNC_RESOLUTION:-1920x1080}"
PENTAGI_URL="${PENTAGI_URL:-https://pentagi:8443}"
VNC_PORT=5901
NOVNC_PORT=6080

log() { echo "[desktop] $*"; }

# ── 1. Start Xvnc ────────────────────────────────────────────────────────────
log "Starting Xvnc on :1 (${RESOLUTION})"
Xvnc :1 \
    -geometry "$RESOLUTION" \
    -depth 24 \
    -rfbport "$VNC_PORT" \
    -SecurityTypes VncAuth \
    -PasswordFile "$HOME/.vnc/passwd" \
    -fg &
XVNC_PID=$!

# Wait until the X display is accepting connections
for i in $(seq 1 20); do
    xdpyinfo -display :1 >/dev/null 2>&1 && break
    sleep 0.5
done
log "Xvnc ready"

# ── 2. Start Xfce4 ───────────────────────────────────────────────────────────
log "Starting Xfce4"
DISPLAY=:1 \
    XDG_RUNTIME_DIR="/tmp/xdg-runtime-root" \
    dbus-launch --exit-with-session startxfce4 &
XFCE_PID=$!

# Give the desktop a moment to paint before launching the browser
sleep 6

# ── 3. Open Chromium pointing at PentAGI ─────────────────────────────────────
log "Opening Chromium → $PENTAGI_URL"
DISPLAY=:1 \
    HOME="$HOME" \
    chromium \
        --no-sandbox \
        --disable-dev-shm-usage \
        --ignore-certificate-errors \
        --start-maximized \
        --disable-infobars \
        --disable-session-crashed-bubble \
        --disable-restore-session-state \
        "$PENTAGI_URL" \
    >/var/log/chromium.log 2>&1 &

# ── 4. Start noVNC (websockify) ───────────────────────────────────────────────
log "Starting noVNC on port $NOVNC_PORT → VNC $VNC_PORT"
websockify \
    --web=/usr/share/novnc/ \
    "$NOVNC_PORT" \
    "localhost:$VNC_PORT" \
    >/var/log/novnc.log 2>&1 &
NOVNC_PID=$!

log "Virtual desktop ready."
log "  noVNC (browser) : http://localhost:$NOVNC_PORT/vnc.html"
log "  VNC direct      : localhost:$VNC_PORT"

# Keep the container alive; restart noVNC if it exits
while true; do
    if ! kill -0 "$XVNC_PID" 2>/dev/null; then
        log "Xvnc exited — shutting down"
        exit 1
    fi
    if ! kill -0 "$NOVNC_PID" 2>/dev/null; then
        log "noVNC exited — restarting"
        websockify \
            --web=/usr/share/novnc/ \
            "$NOVNC_PORT" \
            "localhost:$VNC_PORT" \
            >/var/log/novnc.log 2>&1 &
        NOVNC_PID=$!
    fi
    sleep 5
done
