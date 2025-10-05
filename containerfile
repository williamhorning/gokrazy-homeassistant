FROM alpine:latest

RUN apk add --no-cache bluez dbus udev iproute2 bash

RUN mkdir -p /run/dbus && chmod 755 /run/dbus

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
