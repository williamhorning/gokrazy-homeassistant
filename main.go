package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

const (
	podmanExecutable = "/usr/local/bin/podman"
	bluezImage       = "ghcr.io/williamhorning/gokrazy-homeassistant:latest"
	homeAssistant    = "ghcr.io/home-assistant/home-assistant:dev"
	matterServer     = "ghcr.io/home-assistant-libs/python-matter-server:stable"
)

var sharedArgs = []string{
	"--net=host", "--privileged",
	"-v", "/run/dbus:/run/dbus",
	"-v", "/dev:/dev",
	"-v", "/sys:/sys",
	"--security-opt", "apparmor:unconfined",
	"--cap-add=NET_ADMIN",
	"--cap-add=NET_RAW",
}

func main() {
	slog.Info("starting container setup...")

	stopAllContainers()
	removeAllContainers()
	pullImages()

	createDirs("/perm/homeassistant", "/perm/matter", "/run/dbus")

	startAndStream("bluez", bluezImage, []string{"-v", "/perm/homeassistant:/config", "-v", "/perm/matter:/data/matter"})

	startAndStream("hasst", homeAssistant, []string{"-v", "/perm/homeassistant:/config"})

	startAndStream("mattr", matterServer, []string{"-v", "/perm/matter:/data/matter"})

	waitForSignal()
	stopAllContainers()
}

func podman(args ...string) error {
	cmd := exec.Command(podmanExecutable, args...)
	cmd.Env = expandEnv(os.Environ())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func expandEnv(env []string) []string {
	const extra = "/user:/usr/local/bin"
	for i, val := range env {
		if after, ok := strings.CutPrefix(val, "PATH="); ok {
			env[i] = "PATH=" + extra + ":" + after
			return env
		}
	}
	return append(env, "PATH="+extra+":/usr/local/sbin:/sbin:/usr/sbin:/usr/local/bin:/bin:/usr/bin")
}

func stopAllContainers() {
	slog.Info("stopping existing containers...")
	if err := podman("kill", "--all"); err != nil {
		slog.Warn("couldn't kill containers", "err", err)
	}
}

func removeAllContainers() {
	slog.Info("removing existing containers...")
	if err := podman("rm", "-a", "-f"); err != nil {
		slog.Warn("couldn't remove containers", "err", err)
	}
}

func pullImages() {
	slog.Info("pulling container images...")
	if err := podman("pull", bluezImage, homeAssistant, matterServer); err != nil {
		slog.Warn("failed to pull images", "err", err)
	}
}

func createDirs(dirs ...string) {
	var errs []error
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		slog.Error("failed to create directories", "err", err)
	}
}

func startAndStream(name, image string, extraArgs []string) {
	extraArgs = append(extraArgs, "-e", "PUID=1000", "-e", "PGID=1000", "-e", "TZ=America/New_York")

	if err := startContainer(name, image, extraArgs); err != nil {
		slog.Error("failed to start container", "name", name, "err", err)
		os.Exit(1)
	}
	streamLogs(name)
}

func startContainer(name, image string, extraArgs []string) error {
	args := append([]string{"run", "-d", "--name", name, "--init"}, sharedArgs...)
	args = append(args, extraArgs...)
	args = append(args, image)
	slog.Info("starting container...", "name", name, "image", image)
	return podman(args...)
}

func streamLogs(container string) {
	cmd := exec.Command(podmanExecutable, "logs", "-f", container)
	cmd.Env = append(expandEnv(os.Environ()), "TMPDIR=/tmp")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("failed to get stdout", "container", container, "err", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("failed to get stderr", "container", container, "err", err)
		return
	}

	if err := cmd.Start(); err != nil {
		slog.Error("failed to start logs", "container", container, "err", err)
		return
	}

	go pipeLogs(container, "stdout", stdout)
	go pipeLogs(container, "stderr", stderr)
}

func pipeLogs(container, stream string, reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		msg := scanner.Text()
		switch stream {
		case "stdout":
			slog.Info(fmt.Sprintf("[%s] %s", container, msg))
		case "stderr":
			slog.Warn(fmt.Sprintf("[%s] %s", container, msg))
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Error("log error", "container", container, "stream", stream, "err", err)
	}
}

func waitForSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Warn("received signal, shutting down...", "signal", sig.String())
}
