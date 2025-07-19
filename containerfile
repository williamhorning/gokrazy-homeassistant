FROM debian:stable-slim AS bluez-base

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    bluez \
    dbus-broker \
    udev \
    iproute2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

FROM ghcr.io/home-assistant/home-assistant:dev

COPY --from=bluez-base /sbin/bluetoothd /sbin/
COPY --from=bluez-base /usr/libexec/bluetooth /usr/libexec/bluetooth
COPY --from=bluez-base /etc/bluetooth /etc/bluetooth
COPY --from=bluez-base /usr/bin/dbus-broker /usr/bin/dbus-broker
COPY --from=bluez-base /etc/dbus-1 /etc/dbus-1
COPY --from=bluez-base /lib/udev /lib/udev
COPY --from=bluez-base /usr/bin/udevadm /usr/bin/udevadm

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8123 5580

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
