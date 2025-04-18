package main

import (
	"io"
	"log/slog"
	"strconv"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

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
