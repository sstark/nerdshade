package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type roundFloatTestCase struct {
	in        float64
	precision uint
	expected  float64
}

func TestRoundFloat(t *testing.T) {
	tests := []roundFloatTestCase{
		{0.252341, 2, 0.25},
		{0.25634, 2, 0.26},
		{1.005, 3, 1.005},
		{3.14, 0, 3},
		{3.9, 0, 4},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if result := roundFloat(test.in, test.precision); result != test.expected {
				t.Errorf("Rounded value %f not equal to expected %f", result, test.expected)
			}
		})
	}
}

type TimeRatioTestCase struct {
	from     time.Time
	to       time.Time
	dur      time.Duration
	expected float64
}

func TestTimeRatio(t *testing.T) {
	tests := map[string]TimeRatioTestCase{
		"Equal From To": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Hour,
			1.0,
		},
		"Exactly 100%": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Date(2025, time.April, 15, 18, 27, 0, 0, time.UTC),
			time.Hour,
			0.0,
		},
		"Inside correct interval": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Date(2025, time.April, 15, 17, 58, 0, 0, time.UTC),
			time.Hour,
			0.483,
		},
		"Same times, but longer reference interval": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Date(2025, time.April, 15, 17, 58, 0, 0, time.UTC),
			time.Hour * 2,
			0.742,
		},
		"To is before From": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Date(2025, time.April, 15, 13, 12, 0, 0, time.UTC),
			time.Hour,
			0.0,
		},
		"To is after From, but outside interval": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.UTC),
			time.Date(2025, time.April, 15, 18, 46, 0, 0, time.UTC),
			time.Hour,
			1.0,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			if result := TimeRatio(test.from, test.to, test.dur); result != test.expected {
				t.Errorf("Ratio %f not equal to expected %f", result, test.expected)
			}
		})
	}
}

type BothOrNoneTestCase struct {
	a, b     string
	expected bool
}

func TestBothOrNone(t *testing.T) {
	tests := map[string]BothOrNoneTestCase{
		"none":   {"", "", true},
		"both":   {"foo", "bar", true},
		"only a": {"foo", "", false},
		"only b": {"", "foo", false},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			if result := BothOrNone(test.a, test.b); result != test.expected {
				t.Errorf("Got %v instead of %v", result, test.expected)
			}
		})
	}
}

func TestGetFlags(t *testing.T) {
	t.Run("help is shown", func(t *testing.T) {
		_, output, err := GetFlags("foo", []string{"--help"})
		if err != flag.ErrHelp {
			t.Errorf("Help err wrong. Got %v instead of %v", err, flag.ErrHelp)
		}
		if !strings.HasPrefix(output, "Usage of foo") {
			t.Errorf("Help output wrong. Got\n%v instead of\n%v", output, "Usage of foo")
		}
	})

	t.Run("wakeup/bedtime are used together (unhappy case)", func(t *testing.T) {
		_, _, err := GetFlags("foo", []string{"-fixedWakeup", "10:00"})
		if err.Error() != "both, -fixedBedtime and -fixedWakeup need to be supplied" {
			t.Errorf("Got %v instead of %v", err, "")
		}
	})

	t.Run("wakeup/bedtime are used together (happy case)", func(t *testing.T) {
		_, _, err := GetFlags("foo", []string{"-fixedWakeup", "10:00", "-fixedBedtime", "22:00"})
		if err != nil {
			t.Errorf("Got %v instead of nil", err)
		}
	})
}

func TestGetFlagsConfigFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	config := []byte("tempNight: 3000\nfixedWakeup: \"6:30\"\nfixedBedtime: \"22:00\"\ntransitionDuration: 15m\nV: true\n")
	if err := os.WriteFile(filepath.Join(xdg, "nerdshade.yaml"), config, 0o644); err != nil {
		t.Fatalf("Could not write config file: %v", err)
	}

	t.Run("config file supplies defaults", func(t *testing.T) {
		c, _, err := GetFlags("foo", nil)
		if err != nil {
			t.Fatalf("Got %v instead of nil", err)
		}
		if got := c.GetInt(keyNightTemp); got != 3000 {
			t.Errorf("Got %d instead of 3000", got)
		}
		if got := c.GetString(keyFixedWakeup); got != "6:30" {
			t.Errorf("Got %q instead of \"6:30\"", got)
		}
		if got := c.GetDuration(keyTransitionDuration); got != 15*time.Minute {
			t.Errorf("Got %v instead of 15m", got)
		}
		if c.GetBool(keyVersion) {
			t.Error("Version should not be settable from a config file")
		}
	})

	t.Run("command line overrides config file", func(t *testing.T) {
		c, _, err := GetFlags("foo", []string{"-tempNight", "5000", "-V"})
		if err != nil {
			t.Fatalf("Got %v instead of nil", err)
		}
		if got := c.GetInt(keyNightTemp); got != 5000 {
			t.Errorf("Got %d instead of 5000", got)
		}
		if !c.GetBool(keyVersion) {
			t.Error("Version should be settable on the command line")
		}
	})
}

func TestGetFlagsConfigPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	homeConfig := filepath.Join(home, ".config")
	xdg := t.TempDir()
	writeConfig := func(t *testing.T, dir, tempNight string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("Could not create %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "nerdshade.yaml"), []byte("tempNight: "+tempNight+"\n"), 0o644); err != nil {
			t.Fatalf("Could not write config file: %v", err)
		}
	}
	getNightTemp := func(t *testing.T) int {
		t.Helper()
		c, _, err := GetFlags("foo", nil)
		if err != nil {
			t.Fatalf("Got %v instead of nil", err)
		}
		return c.GetInt(keyNightTemp)
	}
	writeConfig(t, homeConfig, "1111")
	writeConfig(t, xdg, "2222")

	t.Run("XDG_CONFIG_HOME wins over home config dir", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", xdg)
		if got := getNightTemp(t); got != 2222 {
			t.Errorf("Got %d instead of 2222", got)
		}
	})

	t.Run("XDG_CONFIG_HOME replaces the home config dir", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		if got := getNightTemp(t); got != DefaultNightTemp {
			t.Errorf("Got %d instead of %d", got, DefaultNightTemp)
		}
	})

	t.Run("relative XDG_CONFIG_HOME is ignored", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "relative/dir")
		if got := getNightTemp(t); got != 1111 {
			t.Errorf("Got %d instead of 1111", got)
		}
	})

	t.Run("home config dir is used when XDG_CONFIG_HOME is unset", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		if got := getNightTemp(t); got != 1111 {
			t.Errorf("Got %d instead of 1111", got)
		}
	})
}
