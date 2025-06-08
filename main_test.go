package main

import (
	"flag"
	"strconv"
	"strings"
	"testing"
	"time"
)

//func TestMain(t *testing.T) {
//	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
//}

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
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Hour,
			1.0,
		},
		"Exactly 100%": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Date(2025, time.April, 15, 18, 27, 0, 0, time.Local),
			time.Hour,
			0.0,
		},
		"Inside correct interval": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Date(2025, time.April, 15, 17, 58, 0, 0, time.Local),
			time.Hour,
			0.483,
		},
		"Same times, but longer reference interval": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Date(2025, time.April, 15, 17, 58, 0, 0, time.Local),
			time.Hour * 2,
			0.742,
		},
		"To is before From": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Date(2025, time.April, 15, 13, 12, 0, 0, time.Local),
			time.Hour,
			0.0,
		},
		"To is after From, but outside interval": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			time.Date(2025, time.April, 15, 18, 46, 0, 0, time.Local),
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
		if err.Error() != "Both, -fixedBedtime and -fixedWakeup need to be supplied" {
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
