#!/bin/sh
set -e

echo "[entrypoint] starting dbus..."

/usr/bin/dbus-broker --controller --machine-id=default --unix=/run/dbus/system_bus_socket &

DBUS_SOCKET="/run/dbus/system_bus_socket"
timeout=10
while [ ! -S "$DBUS_SOCKET" ] && [ "$timeout" -gt 0 ]; do
  echo "[entrypoint] Waiting for D-Bus socket ($DBUS_SOCKET)..."
  sleep 1
  timeout=$((timeout - 1))
done

if [ ! -S "$DBUS_SOCKET" ]; then
  echo "[entrypoint] dbus not found after timeout: things might fail"
else
  echo "[entrypoint] dbus started"
fi

echo "[entrypoint] starting bluetoothd..."

/sbin/bluetoothd --experimental &

echo "[entrypoint] bluetoothd started"
echo "[entrypoint] handing off..."

exec /usr/local/bin/python3 -m homeassistant --config /config
