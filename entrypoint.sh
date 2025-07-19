#!/bin/sh
set -e

echo "installing tweaked packages..."

pip install --no-cache-dir cheshire==0.5.6 bleak==0.22.3

echo "starting dbus..."

mkdir -p /run/dbus
dbus-broker --controller --unix=/run/dbus/system_bus_socket &

DBUS_SOCKET="/run/dbus/system_bus_socket"
timeout=10
while [ ! -S "$DBUS_SOCKET" ] && [ "$timeout" -gt 0 ]; do
  echo "waiting for ($DBUS_SOCKET)..."
  sleep 1
  timeout=$(("$timeout" - 1))
done

if [ ! -S "$DBUS_SOCKET" ]; then
  echo "dbus not found after timeout, things might fail"
else
  echo "dbus started"
fi

echo "starting bluetooth..."

bluetoothd --experimental &

echo "bluetooth started"
echo "starting matter server..."

mkdir -p /data/matter
python3 -m matter_server.server --storage-path /data/matter --bluetooth-adapter &

echo "matter server started"
echo "starting home assistant..."

python -m homeassistant --config /config
