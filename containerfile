FROM alpine:latest

ENV DEBIAN_FRONTEND=noninteractive

# Install necessary packages
RUN apk update \
    && apk add --no-cache \
    bluez \
    dbus \
    udev \
    iproute2 \
    bash \
    && apk clean

# Create necessary directories for DBus
RUN mkdir -p /run/dbus && chmod 777 /run/dbus

# Copy the entrypoint script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
