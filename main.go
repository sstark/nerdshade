package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"syscall"
	"time"
)

type Config struct {
	Debug              bool
	Help               bool
	NightTemp          int
	DayTemp            int
	NightGamma         int
	DayGamma           int
	Latitude           float64
	Longitude          float64
	Wakeup             string
	Bedtime            string
	WakeupTime         time.Time
	BedtimeTime        time.Time
	Loop               bool
	Version            bool
	HyprctlCmd         string
	TransitionDuration time.Duration
}

const (
	DefaultLatitude           = 48.516
	DefaultLongitude          = 9.120
	DefaultNightTemp          = 4000
	DefaultDayTemp            = 6500
	DefaultNightGamma         = 90
	DefaultDayGamma           = 100
	DefaultLoopInterval       = time.Second * 30
	DefaultTransitionDuration = time.Hour
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
func GetFlags(progname string, args []string) (Config, string, error) {
	c := Config{}
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)
	var out bytes.Buffer
	flags.SetOutput(&out)
	flags.BoolVar(&(c.Debug), "debug", false, "Print debug info")
	flags.IntVar(&(c.NightTemp), "tempNight", DefaultNightTemp, "Night color temperature")
	flags.IntVar(&(c.DayTemp), "tempDay", DefaultDayTemp, "Day color temperature")
	flags.IntVar(&(c.NightGamma), "gammaNight", DefaultNightGamma, "Night gamma")
	flags.IntVar(&(c.DayGamma), "gammaDay", DefaultDayGamma, "Day gamma")
	flags.Float64Var(&(c.Latitude), "latitude", DefaultLatitude, "Your location latitude")
	flags.Float64Var(&(c.Longitude), "longitude", DefaultLongitude, "Your location longitude")
	flags.StringVar(&(c.Wakeup), "fixedWakeup", "", "Wakeup time in 24-hour format, e. g. \"6:00\" (overrides location)")
	flags.StringVar(&(c.Bedtime), "fixedBedtime", "", "Bedtime time in 24-hour format, e. g. \"22:30\" (overrides location)")
	flags.BoolVar(&(c.Loop), "loop", false, "Run nerdshade continuously")
	flags.BoolVar(&(c.Version), "V", false, "Show program version")
	flags.StringVar(&(c.HyprctlCmd), "hyperctl", HyprctlCmd, "Path to hyperctl program")
	flags.DurationVar(&(c.TransitionDuration), "transitionDuration", DefaultTransitionDuration, "Duration of transition, e. g. \"45m\" or \"1h10m\"")
	err := flags.Parse(args)
	if !BothOrNone(c.Wakeup, c.Bedtime) {
		return c, out.String(), errors.New("Both, -fixedBedtime and -fixedWakeup need to be supplied")
	}
	return c, out.String(), err
}

func main() {
	cflags, flagsOut, err := GetFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(flagsOut)
		os.Exit(0)
	}
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
	// TODO: Factor out to a function that returns a useful value to the OS.
	slog.Debug("starting", "localtime", time.Now())
	GetAndSetBrightness(cflags, time.Now())
	if cflags.Loop {
		repeatUntilInterrupt(func() {
			GetAndSetBrightness(cflags, time.Now())
		},
			DefaultLoopInterval,
			syscall.SIGINT,
			syscall.SIGTERM)
	}
}
