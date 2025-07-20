#!/bin/sh
set -e

echo "[entrypoint] setting up dbus..."

mkdir -p /run/dbus
chmod 777 /run/dbus

echo "[entrypoint] starting dbus..."
/usr/bin/dbus-daemon --system --fork

DBUS_SOCKET="/run/dbus/system_bus_socket"
timeout=10
while [ ! -S "$DBUS_SOCKET" ] && [ "$timeout" -gt 0 ]; do
  echo "[entrypoint] waiting for ($DBUS_SOCKET)..."
  sleep 1
  timeout=$((timeout - 1))
done

if [ ! -S "$DBUS_SOCKET" ]; then
  echo "[entrypoint] dbus not found after timeout: things might fail"
else
  echo "[entrypoint] dbus started"
fi

echo "[entrypoint] starting bluetoothd..."
exec /usr/lib/bluetooth/bluetoothd --experimental
