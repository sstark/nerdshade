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
	cmd      string
	key      string
	value    int
	expected string
	errstr   string
}

func TestHyprctl(t *testing.T) {
	tests := map[string]TestHyprctlCase{
		"set temperature": {
			MockHyprctl,
			"temperature",
			5000,
			"subcmd=temperature stdout=\"ok\\n\"",
			"",
		},
		"set gamma": {
			MockHyprctl,
			"gamma",
			95,
			"subcmd=gamma stdout=\"ok\\n\"",
			"",
		},
		"set gamma too high": {
			MockHyprctl,
			"gamma",
			101,
			"subcmd=gamma stdout=\"Invalid gamma value (should be in range 0-100%)\\n\"",
			"",
		},
		"wrong sub command": {
			MockHyprctl,
			"foo",
			123,
			"subcmd=foo stdout=\"invalid command\\n\"",
			"",
		},
		"empty sub command": {
			MockHyprctl,
			"",
			123,
			"subcmd=\"\" stdout=\"invalid command\\n\"",
			"",
		},
		"hyprctl not found": {
			"./notexisting/binary",
			"foo",
			123,
			"subcmd=foo stderr=\"/bin/sh: ",
			"exit status 127",
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
			err := Hyprctl(test.cmd, test.key, test.value)
			got := logOutput.String()
			if !strings.Contains(got, test.expected) {
				t.Errorf("Output from call to hyprctl did not contain expected %s (got: %s)", test.expected, got)
			}
			var expected_errstr string
			if err != nil {
				expected_errstr = err.Error()
			} else {
				expected_errstr = ""
			}
			if expected_errstr != test.errstr {
				t.Errorf("Call to hyprctl did not return expected error string %s (got: %s)", test.errstr, expected_errstr)
			}
		})
	}
}
