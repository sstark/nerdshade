package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

// BrightnessLevel returns the brightness based on the time given
// ranging from 0.0 to 1.0.
// sunrise and sunset times need to be supplied.
func BrightnessLevel(when, sunrise, sunset time.Time) float64 {
	// Night
	if when.Before(sunrise) || when == sunrise || when.After(sunset) || when == sunset {
		slog.Debug("it is night")
		return 0.0
	}
	// Sunrise
	if when.Before(sunrise.Add(TransitionDuration)) {
		return roundFloat3(TimeRatio(when, sunrise.Add(TransitionDuration), TransitionDuration))
	}
	// Sunset
	if when.After(sunset.Add(-TransitionDuration)) && when.Before(sunset) {
		return roundFloat3(1.0 - TimeRatio(when, sunset, TransitionDuration))
	}
	// Day
	return 1.0
}

// ParseHourMinute takes a string of the form "23:34", parses it as a "kitchen"
// time value and returns the hour and minute.
// Only 24-hour style values are supported.
// The third return value is nil or an error message if parsing failed.
func ParseHourMinute(hourMinute string) (hour int, minute int, err error) {
	split := strings.Split(hourMinute, ":")
	if len(split) != 2 {
		err = fmt.Errorf("Time value malformed, needs to be of the form \"HH:MM\"")
		return
	}
	hour, err = strconv.Atoi(split[0])
	if err != nil {
		return
	}
	minute, err = strconv.Atoi(split[1])
	if err != nil {
		return
	}
	if hour < 0 || hour > 23 {
		err = fmt.Errorf("Hour value (%d) must be >=0 and <=23", hour)
		return
	}
	if minute < 0 || hour > 59 {
		err = fmt.Errorf("Minute value (%d) must be >=0 and <=59", hour)
		return
	}
	slog.Debug("parsed hour/minute", "hour", hour, "minute", minute)
	return
}

// GetLocalBrightness returns the current brightness at given location
func GetLocalBrightness(when time.Time, latitude, longitude float64) float64 {
	rise, set := sunrise.SunriseSunset(latitude, longitude, when.Year(), when.Month(), when.Day())
	slog.Debug("calculated sun times", "sunrise", rise, "sunset", set, "lat", latitude, "lon", longitude)
	return BrightnessLevel(when, rise.Local(), set.Local())
}

// GetScheduledBrightness( returns the current brightness based on hard schedule
// wakeup and bedtime values will be parsed and date-completed.
func GetScheduledBrightness(when time.Time, wakeup, bedtime string) (float64, error) {
	wakeupHour, wakeupMinute, err := ParseHourMinute(wakeup)
	if err != nil {
		return 0.0, err
	}
	bedtimeHour, bedtimeMinute, err := ParseHourMinute(bedtime)
	if err != nil {
		return 0.0, err
	}
	rise := time.Date(when.Year(), when.Month(), when.Day(), wakeupHour, wakeupMinute, 0, 0, when.Location())
	set := time.Date(when.Year(), when.Month(), when.Day(), bedtimeHour, bedtimeMinute, 0, 0, when.Location())
	slog.Debug("scheduled wakeup/bedtime", "rise", rise, "set", set)
	return BrightnessLevel(when, rise, set), nil
}

// ScaleBrightness scales the given brightness value to min/max
// Use this for calculating temperature and gamma values from the brightness level
func ScaleBrightness(brightness float64, min, max int) int {
	return int(((float64(max) - float64(min)) * brightness) + float64(min))
}
