package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Debug       bool
	Help        bool
	NightTemp   int
	DayTemp     int
	NightGamma  int
	DayGamma    int
	Latitude    float64
	Longitude   float64
	Wakeup      string
	Bedtime     string
	WakeupTime  time.Time
	BedtimeTime time.Time
	Loop        bool
	Version     bool
	HyprctlCmd  string
}

const (
	DefaultLatitude     = 48.516
	DefaultLongitude    = 9.120
	DefaultNightTemp    = 4000
	DefaultDayTemp      = 6500
	DefaultNightGamma   = 90
	DefaultDayGamma     = 100
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

func BothOrNone(a, b string) bool {
	return (a != "" && b != "") || (a == "" && b == "")
}

// GetFlags creates and returns a new config object from command line flags
func GetFlags() (Config, error) {
	c := Config{}
	flag.BoolVar(&(c.Debug), "debug", false, "Print debug info")
	flag.IntVar(&(c.NightTemp), "tempNight", DefaultNightTemp, "Night color temperature")
	flag.IntVar(&(c.DayTemp), "tempDay", DefaultDayTemp, "Day color temperature")
	flag.IntVar(&(c.NightGamma), "gammaNight", DefaultNightGamma, "Night gamma")
	flag.IntVar(&(c.DayGamma), "gammaDay", DefaultDayGamma, "Day gamma")
	flag.Float64Var(&(c.Latitude), "latitude", DefaultLatitude, "Your location latitude")
	flag.Float64Var(&(c.Longitude), "longitude", DefaultLongitude, "Your location longitude")
	flag.StringVar(&(c.Wakeup), "fixedWakeup", "", "Wakeup time in 24-hour format, e. g. \"6:00\" (overrides location)")
	flag.StringVar(&(c.Bedtime), "fixedBedtime", "", "Bedtime time in 24-hour format, e. g. \"22:30\" (overrides location)")
	flag.BoolVar(&(c.Loop), "loop", false, "Run nerdshade continuously")
	flag.BoolVar(&(c.Version), "V", false, "Show program version")
	flag.StringVar(&(c.HyprctlCmd), "hyperctl", HyprctlCmd, "Path to hyperctl program")
	flag.Parse()
	if !BothOrNone(c.Wakeup, c.Bedtime) {
		return c, errors.New("Both, -fixedBedtime and -fixedWakeup need to be supplied")
	}
	return c, nil
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
			GetAndSetBrightness(cflags, cl.Now())
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
	cflags, err := GetFlags()
	if cflags.Version {
		fmt.Println(Version)
		os.Exit(0)
	}
	if cflags.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if err != nil {
		slog.Error("Error in flags", "error", err)
		os.Exit(1)
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
