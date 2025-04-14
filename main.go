package main

import (
	"log/slog"
	"math"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

const (
	DefaultLatitude    = 48.516
	DefaultLongitude   = 9.120
	TransitionDuration = time.Hour
)

// Coords returns the latitude and longitude as floats
func (loc Location) Coords() (float64, float64) {
	return loc.Latitude, loc.Longitude
}

// NewLocation returns a new location object
func NewLocation() Location {
	return Location{DefaultLatitude, DefaultLongitude}
}

// roundFloat rounds a float to the given precision
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// TimeRatio calculates "how much" from has reached to, relative to dur
// Examples for dur = 1 hour:
//
//	16:45 - 17:00: 0.750
//	05:00 - 05:00: 1.000
//	06:30 - 06:10: 1.000
//
// The result is in 3 digit precision.
func TimeRatio(from, to time.Time, dur time.Duration) float64 {
	if to.Before(from) {
		return 0.0
	}
	// to is out of bounds, but target is still reached
	if to.After(from.Add(dur)) {
		return 1.0
	}
	timeDiff := to.Sub(from)
	ratio := (dur.Seconds() - timeDiff.Seconds()) / dur.Seconds()
	slog.Debug("timeratio", "timeDiff", timeDiff, "ratio", ratio)
	return math.Min(roundFloat(ratio, 3), 1.0)
}

// BrightnessLevel returns the brightness based on the time given
// ranging from 0.0 to 1.0.
// sunrise and sunset times need to be supplied.
func BrightnessLevel(when, sunrise, sunset time.Time) float64 {
	// Night
	if when.Before(sunrise) || when == sunrise || when.After(sunset) || when == sunset {
		slog.Info("it is night")
		return 0.0
	}
	// Sunrise
	if when.Before(sunrise.Add(TransitionDuration)) {
		return TimeRatio(when, sunrise.Add(TransitionDuration), TransitionDuration)
	}
	// Sunset
	if when.After(sunset.Add(-TransitionDuration)) && when.Before(sunset) {
		return TimeRatio(when, sunset, TransitionDuration)
	}
	// Day
	return 1.0
}

// GetLocalBrightness returns the current brightness at given location
func GetLocalBrightness(when time.Time, location Location) float64 {
	latitude, longitude := location.Coords()
	rise, set := sunrise.SunriseSunset(latitude, longitude, when.Year(), when.Month(), when.Day())
	slog.Debug("calculated sun times", "sunrise", rise, "sunset", set, "location", location)
	return BrightnessLevel(when, rise.Local(), set.Local())
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	now := time.Now()
	location := NewLocation()
	slog.Debug("starting", "localtime", now)
	brightness := GetLocalBrightness(now, location)
	slog.Info("local brightness", "brightness", brightness)
	err := SetHyprsunset(brightness)
	if err != nil {
		slog.Warn("error setting brightness", "err", err)
	}
}
