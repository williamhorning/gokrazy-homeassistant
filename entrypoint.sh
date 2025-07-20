#!/bin/sh
set -e

echo "[entrypoint] starting dbus..."

/usr/bin/dbus-broker --controller --unix=/run/dbus/system_bus_socket &

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

exec /sbin/bluetoothd --experimental
