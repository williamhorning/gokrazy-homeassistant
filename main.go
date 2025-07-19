// package main runs a given docker container for home assistant
package main

import (
	"cmp"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

const containerImage = "ghcr.io/williamhorning/gokrazy-homeassistant:latest"

var containerArgs = []string{
	"--init",
	"--net=host",
	"--privileged",
	"-v", "/run/dbus:/run/dbus",
	"-v", "/dev:/dev",
	"-v", "/sys:/sys",
	"--security-opt", "apparmor:unconfined",
	"--cap-add=NET_ADMIN",
	"--cap-add=NET_RAW",
	// persistent storage
	"-v", "/perm/homeassistant:/config",
	"-v", "/perm/matter:/data/matter",
	// home assistant configuration
	"-e", "PUID=1000",
	"-e", "PGID=1000",
	"-e", "TZ=America/New_York",
}

// add things to the path to make things show up
// source: https://gokrazy.org/packages/docker-containers/
func expandPath(env []string) []string {
	extra := "/user:/usr/local/bin"
	found := false

	for idx, val := range env {
		parts := strings.Split(val, "=")
		if len(parts) < 2 {
			continue // malformed entry
		}

		key := parts[0]
		if key != "PATH" {
			continue
		}

		val := strings.Join(parts[1:], "=")
		env[idx] = key + "=" + extra + ":" + val
		found = true
	}

	if !found {
		env = append(env, "PATH="+extra+":"+"/usr/local/sbin:/sbin:/usr/sbin:/usr/local/bin:/bin:/usr/bin")
	}

	return env
}

// podman executes a podman command with the given arguments.
// source: https://gokrazy.org/packages/docker-containers/
func podman(args ...string) error {
	podmanCmd := exec.Command("/usr/local/bin/podman", args...)
	podmanCmd.Env = expandPath(os.Environ())
	podmanCmd.Env = append(podmanCmd.Env, "TMPDIR=/tmp")
	podmanCmd.Stdin = os.Stdin
	podmanCmd.Stdout = os.Stdout
	podmanCmd.Stderr = os.Stderr

	if err := podmanCmd.Run(); err != nil {
		return fmt.Errorf("podman command %v failed: %w", podmanCmd.Args, err)
	}

	return nil
}

func main() {
	slog.Info("starting container... ", "image", containerImage)

	slog.Info("stopping existing containers...")

	if err := podman("kill", "hasst", "mattr"); err != nil {
		slog.Warn("couldn't kill containers (might not be running)", "err", err)
	}

	if err := podman("rm", "hasst", "mattr"); err != nil {
		slog.Warn("couldn't remove containers (might not exist)", "err", err)
	}

	slog.Info("making directories...")

	if err := cmp.Or(
		os.MkdirAll("/perm/homeassistant", os.ModePerm),
		os.MkdirAll("/perm/matter", os.ModePerm),
		os.MkdirAll("/run/dbus", os.ModePerm),
	); err != nil {
		slog.Warn("failed to make dir", "err", err)
	}

	runArgs := []string{"run", "-d", "--name", "hasst"}
	runArgs = append(runArgs, containerArgs...)
	runArgs = append(runArgs, containerImage)

	slog.Info("starting main image...", "args", runArgs)

	if err := podman(runArgs...); err != nil {
		slog.Error("failed to start container", "err", err)

		os.Exit(1)
	}

	runArgs[3] = "mattr"
	runArgs[len(runArgs)-1] = "ghcr.io/home-assistant-libs/python-matter-server:stable"

	slog.Info("starting matter image...", "args", runArgs)

	if err := podman(runArgs...); err != nil {
		slog.Error("failed to start container", "err", err)

		os.Exit(1)
	}

	slog.Info("started container successfully!")

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	s := <-quitChannel

	slog.Warn("received signal, shutting down...", "signal", s.String())

	slog.Warn("stopping container...")

	if err := podman("stop", "hasst", "mattr"); err != nil {
		slog.Error("failed stopping containers", "err", err)
	} else {
		slog.Info("container stopped")
	}

	slog.Error("exited!")
}
