package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
)

const (
	HyprctlCmd = "hyprctl"
)

// Thanks https://stackoverflow.com/questions/6182369/exec-a-shell-command-in-go
func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Hyprctl calls hyprctl to set either temperatur or gamma
func Hyprctl(cmd, subcmd string, val int) error {
	// Unfortunately hyprctl will not write an error to stderr nor return != 0 if
	// supplied with wrong arguments. In the hope this will change we still
	// check properly.
	slog.Debug("running hyprctl", subcmd, val)
	stdout, stderr, err := Shellout(fmt.Sprintf("%s hyprsunset %s %d", cmd, subcmd, val))
	if stderr != "" {
		slog.Warn("hyprctl", "subcmd", subcmd, "stderr", stderr)
	}
	slog.Debug("hyprctl", "subcmd", subcmd, "stdout", stdout)
	return err
}

// SetHyprsunset contacts a running Hyprland session to set temperature
func SetHyprsunsetTemperature(temperature int) error {
	return Hyprctl(HyprctlCmd, "temperature", temperature)
}

// SetHyprsunset contacts a running Hyprland session to set gamma
func SetHyprsunsetGamma(gamma int) error {
	return Hyprctl(HyprctlCmd, "gamma", gamma)
}
