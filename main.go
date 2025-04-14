package main

import (
	"fmt"
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
//	06:30 - 06:10: 0.000
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
	return math.Min(roundFloat(ratio, 3), 1.0)
}

// BrightnessLevel returns the brightness based on the time given
// ranging from 0.0 to 1.0.
// sunrise and sunset times need to be supplied.
func BrightnessLevel(t, sunrise, sunset time.Time) float64 {
	// Night
	if t.Before(sunrise) || t == sunrise || t.After(sunset) || t == sunset {
		return 0.0
	}
	// Sunrise
	if t.Before(sunrise.Add(TransitionDuration)) {
		return TimeRatio(t, sunrise.Add(TransitionDuration), TransitionDuration)
	}
	// Sunset
	if t.After(sunset.Add(-TransitionDuration)) && t.Before(sunset) {
		return TimeRatio(t, sunset, TransitionDuration)
	}
	// Day
	return 1.0
}

func main() {
	now := time.Now()
	lat, lon := NewLocation().Coords()
	rise, set := sunrise.SunriseSunset(lat, lon, now.Year(), now.Month(), now.Day())
	brightness := BrightnessLevel(now, rise.Local(), set.Local())
	fmt.Printf("brightness: %f\n", brightness)
}
