#!/bin/bash
set -e

# Initialise VNC password from environment
mkdir -p "$HOME/.vnc"
printf '%s' "${VNC_PASSWORD:-pentagi}" | vncpasswd -f > "$HOME/.vnc/passwd"
chmod 600 "$HOME/.vnc/passwd"

# xstartup: launches Xfce4 inside the VNC session
cat > "$HOME/.vnc/xstartup" << 'XSTARTUP'
#!/bin/bash
unset SESSION_MANAGER
unset DBUS_SESSION_BUS_ADDRESS
export XDG_RUNTIME_DIR="/tmp/xdg-runtime-root"
mkdir -p "$XDG_RUNTIME_DIR" && chmod 700 "$XDG_RUNTIME_DIR"
exec dbus-launch --exit-with-session startxfce4
XSTARTUP
chmod +x "$HOME/.vnc/xstartup"

exec /start.sh
