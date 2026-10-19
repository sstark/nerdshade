package main

import (
	"bytes"
	"errors"
	"fmt"
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

func stubSleep(t *testing.T) *[]time.Duration {
	t.Helper()
	previous := sleep
	var slept []time.Duration
	sleep = func(d time.Duration) {
		slept = append(slept, d)
	}
	t.Cleanup(func() { sleep = previous })
	return &slept
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
			wantLog: `level=WARN msg="hyprctl still failing, giving up" subcmd=foo attempts=10 err="exit status 127"`,
			errstr:  "exit status 127",
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			logOutput := captureLogs(t)
			slept := stubSleep(t)
			var gotCommand string
			stubShellout(t, func(command string) shellResult {
				gotCommand = command
				return test.result
			})

			err := Hyprctl(test.cmd, test.key, test.value)

			if gotCommand != test.wantCommand {
				t.Errorf("Got shell command %q instead of %q", gotCommand, test.wantCommand)
			}
			if test.errstr != "" {
				if len(*slept) != hyprctlMaxAttempts-1 {
					t.Errorf("Expected %d retries, got %d", hyprctlMaxAttempts-1, len(*slept))
				}
			} else if len(*slept) != 0 {
				t.Errorf("Expected no retries, got %d", len(*slept))
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

func TestHyprctlRetriesUntilSuccess(t *testing.T) {
	logOutput := captureLogs(t)
	slept := stubSleep(t)
	calls := 0
	stubShellout(t, func(command string) shellResult {
		calls++
		if calls < 3 {
			return shellResult{err: errors.New("exit status 1")}
		}
		return shellResult{stdout: "ok\n"}
	})

	err := Hyprctl("hyprctl", "temperature", 5000)

	if err != nil {
		t.Errorf("Expected no error after successful retry, got %v", err)
	}
	if calls != 3 {
		t.Errorf("Expected 3 calls to hyprctl, got %d", calls)
	}
	if len(*slept) != 2 {
		t.Errorf("Expected 2 sleeps, got %d", len(*slept))
	}
	for _, d := range *slept {
		if d != hyprctlRetryInterval {
			t.Errorf("Expected retry interval %v, got %v", hyprctlRetryInterval, d)
		}
	}
	if got := logOutput.String(); strings.Contains(got, "giving up") {
		t.Errorf("Expected no give up message, got: %s", got)
	}
}

func commandsFor(cmd, subcmd string, value, n int) []string {
	var commands []string
	for i := 0; i < n; i++ {
		commands = append(commands, fmt.Sprintf("%s hyprsunset %s %d", cmd, subcmd, value))
	}
	return commands
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
			wantCommands: append(
				commandsFor("./notexisting/binary", "temperature", 6250, hyprctlMaxAttempts),
				commandsFor("./notexisting/binary", "gamma", 99, hyprctlMaxAttempts)...,
			),
			wantLogs: []string{
				`level=WARN msg=hyprctl subcmd=temperature stderr="./notexisting/binary: No such file or directory\n"`,
				`level=WARN msg="hyprctl still failing, giving up" subcmd=temperature attempts=10 err="exit status 127"`,
				`level=WARN msg="error setting temperature" err="exit status 127"`,
				`level=WARN msg=hyprctl subcmd=gamma stderr="./notexisting/binary: No such file or directory\n"`,
				`level=WARN msg="hyprctl still failing, giving up" subcmd=gamma attempts=10 err="exit status 127"`,
				`level=WARN msg="error setting gamma" err="exit status 127"`,
			},
		},
	}
	cflags := newConfig()
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			logOutput := captureLogs(t)
			stubSleep(t)
			var gotCommands []string
			stubShellout(t, func(command string) shellResult {
				gotCommands = append(gotCommands, command)
				return test.result
			})
			cflags.Set(keyHyprctlCmd, test.cmd)

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
