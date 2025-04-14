package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
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

// SetHyprsunset contacts a running Hyprland session by calling hyprctl
func SetHyprsunset(brightness float64) error {
	// Unfortunately hyprctl will not write an error to stderr nor return != 0 if
	// supplied with wrong arguments. In the hope this will change we still
	// check properly.
	stdout, stderr, err := Shellout(fmt.Sprintf("hyprctl hyprsunset temperature %s", "4000"))
	if stderr != "" {
		slog.Warn("hyprctl", "stderr", stderr)
	}
	slog.Debug("hyprctl", "stdout", stdout)
	return err
}
