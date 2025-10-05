# Home Assistant, Matter, and Bluetooth with gokrazy

This repo provides a little program that uses Podman and Go to run Home
Assistant and other things on top of gokrazy.

## Features

- Home Assistant: runs the latest dev build of Home Assistant
- Supports Bluetooth: the container image used includes bluez and dbus-broker
  for things that rely on bluetooth
- Matter support: python-matter-server is also included for matter devices
- Persistent storage: this saves data automatically to your /perm partition

## Add it

Add the following packages to gokrazy, then build and deploy as usual:

```sh
# podman support
gok add github.com/gokrazy/iptables
gok add github.com/gokrazy/nsenter
gok add github.com/gokrazy/podman
gok add github.com/greenpau/cni-plugins/cmd/cni-nftables-portmap
gok add github.com/greenpau/cni-plugins/cmd/cni-nftables-firewall
# home assistant
gok add github.com/williamhorning/gokrazy-homeassistant
```

Home Assistant will be on port 8123 and Matter will use port 5580

## License

The very little code in this repo is under the MIT License
