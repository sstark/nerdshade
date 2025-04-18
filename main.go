package main

import (
	"flag"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

type Config struct {
	Debug     bool
	Help      bool
	MinTemp   int
	MaxTemp   int
	Latitude  float64
	Longitude float64
	Loop      bool
}

const (
	DefaultLatitude     = 48.516
	DefaultLongitude    = 9.120
	DefaultMinTemp      = 4000
	DefaultMaxTemp      = 6500
	// Should be reasonably small to allow for a smooth transition
	DefaultLoopInterval = time.Second * 30
	TransitionDuration  = time.Hour
)

// roundFloat rounds a float to the given precision
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func roundFloat3(val float64) float64 {
	return roundFloat(val, 3)
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
	return math.Min(roundFloat3(ratio), 1.0)
}

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

// GetLocalBrightness returns the current brightness at given location, as well as
// duration until when the next change is expected.
func GetLocalBrightness(when time.Time, latitude, longitude float64) (float64, time.Duration) {
	var rise, set time.Time
	todayRise, todaySet := sunrise.SunriseSunset(latitude, longitude, when.Year(), when.Month(), when.Day())
	if when.After(todaySet) {
		// Get tomorrows rise/set in case we are past sunset already
		rise, set = sunrise.SunriseSunset(latitude, longitude, when.Year(), when.Month(), when.Add(time.Hour * 24).Day())
	} else {
		rise, set = todayRise, todaySet
	}
	slog.Debug("calculated sun times", "sunrise", rise, "sunset", set, "lat", latitude, "lon", longitude)
	level := BrightnessLevel(when, rise.Local(), set.Local())
	// Calculate times to wait for the next transition with a generous buffer of at
	// least one default interval. We do not want to miss the beginning of a transition
	// because the clock was off for a few seconds.
	bufferBeforeTransition := (DefaultLoopInterval + time.Second * 5)
	switch {
	case level == 0.0:
		{
			slog.Debug("waiting for sunrise", "buffer", bufferBeforeTransition)
			return level, rise.Sub(when) - bufferBeforeTransition
		}
	case level == 1.0:
		{
			slog.Debug("waiting for sunset", "buffer", bufferBeforeTransition)
			return level, set.Sub(when.Add(-TransitionDuration)) - bufferBeforeTransition
		}
	}
	return level, DefaultLoopInterval
}

func BrightnessToTemperature(brightness float64, min, max int) int {
	return int(((float64(max) - float64(min)) * brightness) + float64(min))
}

// GetFlags creates and returns a new config object from command line flags
func GetFlags() Config {
	c := Config{}
	flag.BoolVar(&(c.Debug), "debug", false, "Print debug info")
	flag.IntVar(&(c.MinTemp), "min", DefaultMinTemp, "Minimum color temperature")
	flag.IntVar(&(c.MaxTemp), "max", DefaultMaxTemp, "Maximum color temperature")
	flag.Float64Var(&(c.Latitude), "latitude", DefaultLatitude, "Your location latitude")
	flag.Float64Var(&(c.Longitude), "longitude", DefaultLongitude, "Your location longitude")
	flag.BoolVar(&(c.Loop), "loop", false, "Run nerdshade continuously")
	flag.Parse()
	return c
}

func GetAndSetBrightness(cflags Config, when time.Time) time.Duration {
	brightness, waitTime := GetLocalBrightness(when, cflags.Latitude, cflags.Longitude)
	slog.Debug("local brightness", "brightness", brightness)
	err := SetHyprsunset(BrightnessToTemperature(brightness, cflags.MinTemp, cflags.MaxTemp))
	if err != nil {
		slog.Warn("error setting brightness", "err", err)
	}
	return waitTime
}

func MainLoop(cflags Config, cl clock) {
	slog.Info("running continuously")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)
	signal.Notify(sigc, syscall.SIGINT)
	ticker := time.NewTicker(DefaultLoopInterval)
	slog.Debug("loop timing", "interval", DefaultLoopInterval)
	quit := make(chan bool)
	GetAndSetBrightness(cflags, cl.Now())
	for {
		select {
		case <-ticker.C:
			// TODO: use ticker.Reset(interval) here to dynamically adjust ticker
			// interval. We only need to wake up when needed.
			// GetAndSetBrightness could simply return a time.Duration until the next
			// expected change.
			timeWait := GetAndSetBrightness(cflags, cl.Now())
			slog.Debug("wait for next transition", "wait", timeWait)
			ticker.Reset(timeWait)
		case sig := <-sigc:
			slog.Debug("main loop received signal", "signal", sig)
			go func() { quit <- true }()
		case <-quit:
			ticker.Stop()
			slog.Debug("main loop quit")
			return
		}
	}
}

func main() {
	cflags := GetFlags()
	if cflags.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	cl := new(realClock)
	now := cl.Now()
	// TODO: Factor out to a function that returns a useful value to the OS.
	slog.Debug("starting", "localtime", now)
	if cflags.Loop {
		MainLoop(cflags, cl)
	} else {
		GetAndSetBrightness(cflags, now)
	}
}
