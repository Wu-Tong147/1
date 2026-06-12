FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive \
    HOME=/root \
    DISPLAY=:1 \
    VNC_RESOLUTION=1920x1080 \
    VNC_PASSWORD=pentagi \
    PENTAGI_URL=https://pentagi:8443

RUN apt-get update && apt-get install -y --no-install-recommends \
    # Desktop environment
    xfce4 \
    xfce4-terminal \
    xfce4-taskmanager \
    xfce4-screenshooter \
    # VNC server
    tigervnc-standalone-server \
    tigervnc-common \
    # noVNC web client
    novnc \
    websockify \
    # Browser
    chromium \
    # X11 utilities
    dbus-x11 \
    x11-xserver-utils \
    x11-utils \
    xdotool \
    # Fonts
    fonts-liberation \
    fonts-noto-core \
    # Utilities
    procps \
    curl \
    wget \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Suppress xfce4-session warning about dbus
RUN mkdir -p /root/.config/xfce4/xfconf/xfce-perchannel-xml

COPY desktop/entrypoint.sh /entrypoint.sh
COPY desktop/start.sh /start.sh

RUN chmod +x /entrypoint.sh /start.sh

EXPOSE 6080 5901

ENTRYPOINT ["/entrypoint.sh"]
