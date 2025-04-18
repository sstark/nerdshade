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

type TestHyprctlCase struct {
	key   string
	value int
	expected string
}

func TestHyprctl(t *testing.T) {
	tests := map[string]TestHyprctlCase{
		"set temperature": {
			"temperature",
			5000,
			"subcmd=temperature stdout=\"ok\\n\"",
		},
		"set gamma": {
			"gamma",
			95,
			"subcmd=gamma stdout=\"ok\\n\"",
		},
		"set gamma too high": {
			"gamma",
			101,
			"subcmd=gamma stdout=\"Invalid gamma value (should be in range 0-100%)\\n\"",
		},
		"wrong sub command": {
			"foo",
			123,
			"subcmd=foo stdout=\"invalid command\\n\"",
		},
		"empty sub command": {
			"",
			123,
			"subcmd=\"\" stdout=\"invalid command\\n\"",
		},
	}
	logOutput := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			// Start with empty log for each test
			logOutput.Reset()
			Hyprctl(MockHyprctl, test.key, test.value)
			got := logOutput.String()
			if !strings.Contains(got, test.expected) {
				t.Errorf("Output from call to hyprctl did not contain expected %s (got: %s)", test.expected, got)
			}
		})
	}
}
