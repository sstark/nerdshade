package main

import (
	"io"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/nathan-osman/go-sunrise"
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

type BrightnessLevelTestCase struct {
	t        time.Time
	expected float64
}

func TestBrightnessLevel(t *testing.T) {
	tests := map[string]BrightnessLevelTestCase{
		"Before sunset": {
			time.Date(2025, time.April, 15, 17, 27, 0, 0, time.Local),
			1.0,
		},
		"In the middle of sunset": {
			time.Date(2025, time.April, 15, 19, 30, 0, 0, time.Local),
			0.728,
		},
		"Towards the end of sunset": {
			time.Date(2025, time.April, 14, 20, 5, 0, 0, time.Local),
			0.120,
		},
		"Right before end of sunset": {
			time.Date(2025, time.April, 14, 20, 11, 0, 0, time.Local),
			0.020,
		},
		"Right after sunset": {
			time.Date(2025, time.April, 15, 20, 14, 0, 0, time.Local),
			0.0,
		},
		"Right before sunrise": {
			time.Date(2025, time.April, 16, 6, 15, 0, 0, time.Local),
			0.0,
		},
		"Sun has almost risen": {
			time.Date(2025, time.April, 16, 7, 25, 0, 0, time.Local),
			0.900,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			rise, set := sunrise.SunriseSunset(DefaultLatitude, DefaultLongitude, test.t.Year(), test.t.Month(), test.t.Day())
			if result := BrightnessLevel(test.t, rise, set); result != test.expected {
				// Additional logging to make it easier to spot rounding issues
				t.Log(result)
				t.Log(test.expected)
				t.Errorf("Brightness level %f not equal to expected %f", result, test.expected)
			}
		})
	}
}

func TestGetLocalBrightness(t *testing.T) {
	tests := map[string]BrightnessLevelTestCase{
		"In the middle of sunset": {
			time.Date(2025, time.April, 15, 19, 30, 0, 0, time.Local),
			0.728,
		},
		"Right after sunset": {
			time.Date(2025, time.April, 15, 20, 14, 0, 0, time.Local),
			0.0,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			if result := GetLocalBrightness(test.t, DefaultLatitude, DefaultLongitude); result != test.expected {
				t.Errorf("Brightness level %f not equal to expected %f", result, test.expected)
			}
		})
	}
}

type ScaleBrightnessTestcase struct {
	brightness float64
	min int
	max int
	expected int
}

func TestScaleBrightness(t *testing.T) {
	tests := map[string]ScaleBrightnessTestcase{
		"Min Temp": {
			0.0,
			DefaultNightTemp,
			DefaultDayTemp,
			DefaultNightTemp,
		},
		"Medium Temp": {
			0.45,
			DefaultNightTemp,
			DefaultDayTemp,
			5125,
		},
		"Max Temp": {
			1.0,
			DefaultNightTemp,
			DefaultDayTemp,
			DefaultDayTemp,
		},
		"Min Gamma": {
			1.0,
			DefaultNightGamma,
			DefaultDayGamma,
			DefaultDayGamma,
		},
		"Medium Gamma": {
			0.32,
			DefaultNightGamma,
			DefaultDayGamma,
			93,
		},
		"Max Gamma": {
			1.0,
			DefaultNightGamma,
			DefaultDayGamma,
			DefaultDayGamma,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			if result := ScaleBrightness(test.brightness, test.min, test.max); result != test.expected {
				t.Errorf("Mapping to temperature %d not equal to expected %d", result, test.expected)
			}
		})
	}
}
