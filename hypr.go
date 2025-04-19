package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"time"
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
func SetHyprsunsetTemperature(cflags Config, temperature int) error {
	return Hyprctl(cflags.HyprctlCmd, "temperature", temperature)
}

// SetHyprsunset contacts a running Hyprland session to set gamma
func SetHyprsunsetGamma(cflags Config, gamma int) error {
	return Hyprctl(cflags.HyprctlCmd, "gamma", gamma)
}

// GetAndSetBrightness gets the brightness, gets scaled values for temperature
// and gamma and sets those in hyprland.
func GetAndSetBrightness(cflags Config, when time.Time) {
	brightness, err := GetBrightness(cflags, when)
	if err != nil {
		slog.Warn("error getting brightness", "err", err)
	}
	newTemperature := ScaleBrightness(brightness, cflags.NightTemp, cflags.DayTemp)
	newGamma := ScaleBrightness(brightness, cflags.NightGamma, cflags.DayGamma)
	err = SetHyprsunsetTemperature(cflags, newTemperature)
	if err != nil {
		slog.Warn("error setting temperature", "err", err)
	}
	err = SetHyprsunsetGamma(cflags, newGamma)
	if err != nil {
		slog.Warn("error setting gamma", "err", err)
	}
}
