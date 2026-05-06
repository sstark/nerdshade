package main

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type shellResult struct {
	stdout string
	stderr string
	err    error
}

func stubShellout(t *testing.T, fn func(command string) shellResult) {
	t.Helper()
	previous := shellout
	shellout = func(command string) (string, string, error) {
		result := fn(command)
		return result.stdout, result.stderr, result.err
	}
	t.Cleanup(func() { shellout = previous })
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	logOutput := new(bytes.Buffer)
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return logOutput
}

type TestHyprctlCase struct {
	cmd         string
	key         string
	value       int
	wantCommand string
	result      shellResult
	wantLog     string
	errstr      string
}

func TestHyprctl(t *testing.T) {
	tests := map[string]TestHyprctlCase{
		"set temperature": {
			cmd:         "hyprctl",
			key:         "temperature",
			value:       5000,
			wantCommand: "hyprctl hyprsunset temperature 5000",
			result:      shellResult{stdout: "ok\n"},
			wantLog:     `msg=hyprctl subcmd=temperature stdout="ok\n"`,
			errstr:      "",
		},
		"set gamma": {
			cmd:         "hyprctl",
			key:         "gamma",
			value:       95,
			wantCommand: "hyprctl hyprsunset gamma 95",
			result:      shellResult{stdout: "ok\n"},
			wantLog:     `msg=hyprctl subcmd=gamma stdout="ok\n"`,
			errstr:      "",
		},
		"set gamma too high": {
			cmd:         "hyprctl",
			key:         "gamma",
			value:       101,
			wantCommand: "hyprctl hyprsunset gamma 101",
			result:      shellResult{stdout: "Invalid gamma value (should be in range 0-100%)\n"},
			wantLog:     `msg=hyprctl subcmd=gamma stdout="Invalid gamma value (should be in range 0-100%)\n"`,
			errstr:      "",
		},
		"wrong sub command": {
			cmd:         "hyprctl",
			key:         "foo",
			value:       123,
			wantCommand: "hyprctl hyprsunset foo 123",
			result:      shellResult{stdout: "invalid command\n"},
			wantLog:     `msg=hyprctl subcmd=foo stdout="invalid command\n"`,
			errstr:      "",
		},
		"empty sub command": {
			cmd:         "hyprctl",
			key:         "",
			value:       123,
			wantCommand: "hyprctl hyprsunset  123",
			result:      shellResult{stdout: "invalid command\n"},
			wantLog:     `msg=hyprctl subcmd="" stdout="invalid command\n"`,
			errstr:      "",
		},
		"hyprctl not found": {
			cmd:         "./notexisting/binary",
			key:         "foo",
			value:       123,
			wantCommand: "./notexisting/binary hyprsunset foo 123",
			result: shellResult{
				stderr: "./notexisting/binary: No such file or directory\n",
				err:    errors.New("exit status 127"),
			},
			wantLog: `stderr="./notexisting/binary: No such file or directory\n"`,
			errstr:  "exit status 127",
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			logOutput := captureLogs(t)
			var gotCommand string
			stubShellout(t, func(command string) shellResult {
				gotCommand = command
				return test.result
			})

			err := Hyprctl(test.cmd, test.key, test.value)

			if gotCommand != test.wantCommand {
				t.Errorf("Got shell command %q instead of %q", gotCommand, test.wantCommand)
			}
			got := logOutput.String()
			if !strings.Contains(got, test.wantLog) {
				t.Errorf("Output from call to hyprctl did not contain expected %s (got: %s)", test.wantLog, got)
			}
			var gotErrstr string
			if err != nil {
				gotErrstr = err.Error()
			}
			if gotErrstr != test.errstr {
				t.Errorf("Call to hyprctl did not return expected error string %s (got: %s)", test.errstr, gotErrstr)
			}
		})
	}
}

type TestGetAndSetBrightnessCase struct {
	cmd          string
	when         time.Time
	result       shellResult
	wantCommands []string
	wantLogs     []string
}

func TestGetAndSetBrightness(t *testing.T) {
	tests := map[string]TestGetAndSetBrightnessCase{
		"OK case": {
			cmd:    "hyprctl",
			when:   time.Date(2025, time.April, 16, 5, 25, 0, 0, time.UTC),
			result: shellResult{stdout: "ok\n"},
			wantCommands: []string{
				"hyprctl hyprsunset temperature 6250",
				"hyprctl hyprsunset gamma 99",
			},
			wantLogs: []string{
				`msg="local brightness" brightness=0.9`,
				`msg=hyprctl subcmd=temperature stdout="ok\n"`,
				`msg=hyprctl subcmd=gamma stdout="ok\n"`,
			},
		},
		"Not OK case": {
			cmd:  "./notexisting/binary",
			when: time.Date(2025, time.April, 16, 5, 25, 0, 0, time.UTC),
			result: shellResult{
				stderr: "./notexisting/binary: No such file or directory\n",
				err:    errors.New("exit status 127"),
			},
			wantCommands: []string{
				"./notexisting/binary hyprsunset temperature 6250",
				"./notexisting/binary hyprsunset gamma 99",
			},
			wantLogs: []string{
				`level=WARN msg=hyprctl subcmd=temperature stderr="./notexisting/binary: No such file or directory\n"`,
				`level=WARN msg="error setting temperature" err="exit status 127"`,
				`level=WARN msg="error setting gamma" err="exit status 127"`,
			},
		},
	}
	cflags := Config{}
	cflags.HyprctlCmd = ""
	cflags.Latitude = DefaultLatitude
	cflags.Longitude = DefaultLongitude
	cflags.DayGamma = DefaultDayGamma
	cflags.NightGamma = DefaultNightGamma
	cflags.DayTemp = DefaultDayTemp
	cflags.NightTemp = DefaultNightTemp
	cflags.TransitionDuration = DefaultTransitionDuration
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			logOutput := captureLogs(t)
			var gotCommands []string
			stubShellout(t, func(command string) shellResult {
				gotCommands = append(gotCommands, command)
				return test.result
			})
			cflags.HyprctlCmd = test.cmd

			GetAndSetBrightness(cflags, test.when)

			if len(gotCommands) != len(test.wantCommands) {
				t.Fatalf("Expected shell commands %v, got %v", test.wantCommands, gotCommands)
			}
			for i, want := range test.wantCommands {
				if gotCommands[i] != want {
					t.Errorf("Expected shell command %q, got %q", want, gotCommands[i])
				}
			}
			got := logOutput.String()
			for _, want := range test.wantLogs {
				if !strings.Contains(got, want) {
					t.Errorf("Output from call to hyprctl did not contain expected %s (got: %s)", want, got)
				}
			}
		})
	}
}
