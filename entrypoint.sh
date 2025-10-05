#!/bin/sh
set -e

echo "[entrypoint] Starting DBus..."

mkdir -p /run/dbus
chmod 755 /run/dbus

/usr/bin/dbus-daemon --system --fork

DBUS_SOCKET="/run/dbus/system_bus_socket"
for i in $(seq 1 10); do [ -S "$DBUS_SOCKET" ] && break
  echo "[entrypoint] Waiting for DBus socket..."
  sleep 1
done

if [ ! -S "$DBUS_SOCKET" ]; then
  echo "[entrypoint] DBus socket not found after timeout"
else
  echo "[entrypoint] DBus is up"
fi

echo "[entrypoint] Starting bluetoothd..."
exec /usr/lib/bluetooth/bluetoothd --experimental
