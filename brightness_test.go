package main

import (
	"errors"
	"testing"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

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
			if result := BrightnessLevel(test.t, rise, set, DefaultTransitionDuration); result != test.expected {
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
			if result := GetLocalBrightness(test.t, DefaultLatitude, DefaultLongitude, DefaultTransitionDuration); result != test.expected {
				t.Errorf("Brightness level %f not equal to expected %f", result, test.expected)
			}
		})
	}
}

type ScaleBrightnessTestcase struct {
	brightness float64
	min        int
	max        int
	expected   int
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

type ParseHourMinuteTestcase struct {
	s                              string
	expected_hour, expected_minute int
	expected_err                   bool
}

func TestParseHourMinute(t *testing.T) {
	tests := map[string]ParseHourMinuteTestcase{
		"valid format within bounds 1": {
			"14:23", 14, 23, false,
		},
		"valid format within bounds 2": {
			"04:14", 4, 14, false,
		},
		"valid format within bounds 3": {
			"9:00", 9, 0, false,
		},
		"valid format, minute out of bounds": {
			"23:60", 23, 60, true,
		},
		"valid format, hour out of bounds": {
			"24:20", 24, 20, true,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			hour, minute, err := ParseHourMinute(test.s)
			if hour != test.expected_hour {
				t.Errorf("Got hour %d instead of expected %d for %s", hour, test.expected_hour, test.s)
			}
			if minute != test.expected_minute {
				t.Errorf("Got minute %d instead of expected %d for %s", minute, test.expected_minute, test.s)
			}
			if (err != nil) != test.expected_err {
				t.Errorf("Error expected? (%v) but got %v", test.expected_err, err)
			}
		})
	}
}

type ScheduledBrightnessLevelTestCase struct {
	t        time.Time
	wakeup   string
	bedtime  string
	expected float64
	err      error
}

func TestGetScheduledBrightness(t *testing.T) {
	tests := map[string]ScheduledBrightnessLevelTestCase{
		"Before wakeup": {
			time.Date(2025, time.April, 15, 7, 27, 0, 0, time.Local),
			"10:00",
			"22:00",
			0.0,
			nil,
		},
		"In the middle of wakeup": {
			time.Date(2025, time.April, 15, 8, 30, 0, 0, time.Local),
			"8:00",
			"22:00",
			0.500,
			nil,
		},
		"Towards the end of wakeup": {
			time.Date(2025, time.April, 14, 8, 55, 0, 0, time.Local),
			"8:00",
			"22:59",
			0.917,
			nil,
		},
		"Right before bedtime": {
			time.Date(2025, time.April, 14, 20, 11, 0, 0, time.Local),
			"7:00",
			"21:30",
			1.0,
			nil,
		},
		"In the middle of bedtime": {
			time.Date(2025, time.April, 15, 20, 14, 0, 0, time.Local),
			"7:21",
			"20:45",
			0.517,
			nil,
		},
		"After bedtime": {
			time.Date(2025, time.April, 16, 23, 15, 0, 0, time.Local),
			"6:30",
			"21:00",
			0.0,
			nil,
		},
		"After bedtime with bedtime parsing error": {
			time.Date(2025, time.April, 16, 23, 15, 0, 0, time.Local),
			"6:30",
			"21:70",
			0.0,
			errors.New("Minute value (70) must be >=0 and <=59"),
		},
		"After bedtime with wakeup parsing error": {
			time.Date(2025, time.April, 16, 23, 15, 0, 0, time.Local),
			"630",
			"21:10",
			0.0,
			errors.New("Time value malformed, needs to be of the form \"HH:MM\""),
		},
		"Wrong hour format": {
			time.Date(2025, time.April, 16, 23, 15, 0, 0, time.Local),
			"9:00",
			"ab:00",
			0.0,
			errors.New("strconv.Atoi: parsing \"ab\": invalid syntax"),
		},
		"Wrong minute format": {
			time.Date(2025, time.April, 16, 23, 15, 0, 0, time.Local),
			"9:00",
			"21:cd",
			0.0,
			errors.New("strconv.Atoi: parsing \"cd\": invalid syntax"),
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			result, err := GetScheduledBrightness(test.t, test.wakeup, test.bedtime, time.Hour)
			if result != test.expected {
				// Additional logging to make it easier to spot rounding issues
				t.Log(result)
				t.Log(test.expected)
				t.Errorf("Brightness level %f not equal to expected %f", result, test.expected)
			}
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Error expected? (%v) but got %v", test.err, err)
			}
		})
	}
}
