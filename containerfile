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

FROM ghcr.io/home-assistant/home-assistant:dev AS homeassistant-custom

FROM ghcr.io/home-assistant-libs/python-matter-server:stable AS matter-server-base

ARG MATTER_SITE_PACKAGES
RUN export MATTER_SITE_PACKAGES=$(python3 -c "import site; print(site.getsitepackages()[0])") \
    && echo "$MATTER_SITE_PACKAGES" > /tmp/matter_site_packages_path

FROM homeassistant-custom

COPY --from=bluez-base /sbin/bluetoothd /sbin/
COPY --from=bluez-base /usr/libexec/bluetooth /usr/libexec/bluetooth
COPY --from=bluez-base /etc/bluetooth /etc/bluetooth
COPY --from=bluez-base /usr/bin/dbus-broker /usr/bin/dbus-broker
COPY --from=bluez-base /etc/dbus-1 /etc/dbus-1
COPY --from=bluez-base /lib/udev /lib/udev
COPY --from=bluez-base /usr/bin/udevadm /usr/bin/udevadm

ARG HA_SITE_PACKAGES
RUN export HA_SITE_PACKAGES=$(python3 -c "import site; print(site.getsitepackages()[0])") \
    && echo "$HA_SITE_PACKAGES"

RUN export MATTER_SITE_PACKAGES=$(cat /tmp/matter_site_packages_path) && \
    echo "Matter site packages path: $MATTER_SITE_PACKAGES"

COPY --from=matter-server-base ${MATTER_SITE_PACKAGES} ${HA_SITE_PACKAGES}

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8123 5580
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
