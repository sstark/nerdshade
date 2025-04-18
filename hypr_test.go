package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

const (
	MockHyprctl = "./mock_hyprctl.sh"
)

type TestHyprctlOKCase struct {
	key   string
	value int
}

func TestHyprctlOK(t *testing.T) {
	tests := map[string]TestHyprctlOKCase{
		"set temperature": {
			"temperature",
			5000,
		},
		"set gamma": {
			"gamma",
			95,
		},
	}
	logStdout := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(logStdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			Hyprctl(MockHyprctl, test.key, test.value)
			got := logStdout.String()
			expected := "stdout=\"ok\\n\""
			if !strings.Contains(got, expected) {
				t.Errorf("Call to hyprctl did not contain expected %s (got: %s)", expected, got)
			}
		})
	}
}
