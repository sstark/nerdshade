package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

// Config is the viper instance that holds all configuration values.
type Config = *viper.Viper

const (
	keyDebug              = "debug"
	keyNoLogTime          = "nologtime"
	keyNightTemp          = "tempNight"
	keyDayTemp            = "tempDay"
	keyNightGamma         = "gammaNight"
	keyDayGamma           = "gammaDay"
	keyLatitude           = "latitude"
	keyLongitude          = "longitude"
	keyFixedWakeup        = "fixedWakeup"
	keyFixedBedtime       = "fixedBedtime"
	keyLoop               = "loop"
	keyVersion            = "V"
	keyHyprctlCmd         = "hyperctl"
	keyTransitionDuration = "transitionDuration"
)

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

// newConfig returns a Config with all defaults set, without reading a config file.
func newConfig() Config {
	v := viper.New()
	v.SetDefault(keyNightTemp, DefaultNightTemp)
	v.SetDefault(keyDayTemp, DefaultDayTemp)
	v.SetDefault(keyNightGamma, DefaultNightGamma)
	v.SetDefault(keyDayGamma, DefaultDayGamma)
	v.SetDefault(keyLatitude, DefaultLatitude)
	v.SetDefault(keyLongitude, DefaultLongitude)
	v.SetDefault(keyHyprctlCmd, HyprctlCmd)
	v.SetDefault(keyTransitionDuration, DefaultTransitionDuration)
	return v
}

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
	// `to` is out of bounds, but `target` is still reached
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
	// Throwaway buffer, only there because the flag package needs storage for
	// parsed values. The returned Config is the source of truth.
	var f struct {
		Debug              bool
		NoLogTime          bool
		NightTemp          int
		DayTemp            int
		NightGamma         int
		DayGamma           int
		Latitude           float64
		Longitude          float64
		Wakeup             string
		Bedtime            string
		Loop               bool
		Version            bool
		HyprctlCmd         string
		TransitionDuration time.Duration
	}
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)
	var out bytes.Buffer
	flags.SetOutput(&out)
	v := newConfig()
	v.SetConfigName("nerdshade")
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" && filepath.IsAbs(xdg) {
		v.AddConfigPath(xdg)
	} else {
		v.AddConfigPath("$HOME/.config")
	}
	_ = v.ReadInConfig()
	flags.BoolVar(&(f.Debug), keyDebug, v.GetBool(keyDebug), "Print debug info")
	flags.BoolVar(&(f.NoLogTime), keyNoLogTime, v.GetBool(keyNoLogTime), "Omit time from log output (e. g. when running as systemd unit)")
	flags.IntVar(&(f.NightTemp), keyNightTemp, v.GetInt(keyNightTemp), "Night color temperature")
	flags.IntVar(&(f.DayTemp), keyDayTemp, v.GetInt(keyDayTemp), "Day color temperature")
	flags.IntVar(&(f.NightGamma), keyNightGamma, v.GetInt(keyNightGamma), "Night gamma")
	flags.IntVar(&(f.DayGamma), keyDayGamma, v.GetInt(keyDayGamma), "Day gamma")
	flags.Float64Var(&(f.Latitude), keyLatitude, v.GetFloat64(keyLatitude), "Your location latitude")
	flags.Float64Var(&(f.Longitude), keyLongitude, v.GetFloat64(keyLongitude), "Your location longitude")
	flags.StringVar(&(f.Wakeup), keyFixedWakeup, v.GetString(keyFixedWakeup), "Wakeup time in 24-hour format, e. g. \"6:00\" (overrides location)")
	flags.StringVar(&(f.Bedtime), keyFixedBedtime, v.GetString(keyFixedBedtime), "Bedtime time in 24-hour format, e. g. \"22:30\" (overrides location)")
	flags.BoolVar(&(f.Loop), keyLoop, v.GetBool(keyLoop), "Run nerdshade continuously")
	flags.BoolVar(&(f.Version), keyVersion, false, "Show program version")
	flags.StringVar(&(f.HyprctlCmd), keyHyprctlCmd, v.GetString(keyHyprctlCmd), "Path to hyperctl program")
	flags.DurationVar(&(f.TransitionDuration), keyTransitionDuration, v.GetDuration(keyTransitionDuration), "Duration of transition, e. g. \"45m\" or \"1h10m\"")
	err := flags.Parse(args)
	// Command line flags take precedence over the config file.
	flags.Visit(func(fl *flag.Flag) { v.Set(fl.Name, fl.Value.String()) })
	// -V only makes sense on the command line, never from a config file.
	v.Set(keyVersion, f.Version)
	if !BothOrNone(f.Wakeup, f.Bedtime) {
		return v, out.String(), errors.New("both, -fixedBedtime and -fixedWakeup need to be supplied")
	}
	return v, out.String(), err
}

func mainLoop(cflags Config) int {
	slog.Debug("starting", "localtime", time.Now())
	doit := func() {
		GetAndSetBrightness(cflags, time.Now())
	}
	doit()
	if cflags.GetBool(keyLoop) {
		repeatUntilInterrupt(doit, DefaultLoopInterval, syscall.SIGINT, syscall.SIGTERM)
	}
	return 0
}

func main() {
	cflags, flagsOut, err := GetFlags(os.Args[0], os.Args[1:])
	debugLevel := slog.LevelInfo
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println(flagsOut)
		os.Exit(0)
	}
	if cflags.GetBool(keyVersion) {
		fmt.Println(Version)
		os.Exit(0)
	}
	if cflags.GetBool(keyDebug) {
		debugLevel = slog.LevelDebug
		slog.SetLogLoggerLevel(debugLevel)
		cfile := cflags.ConfigFileUsed()
		if cfile != "" {
			slog.Debug("using", "configfile", cfile)
		}
	}
	if cflags.GetBool(keyNoLogTime) {
		slog.SetDefault(slog.New(noTimeHandler(debugLevel)))
	}
	if err != nil {
		slog.Error("Error in flags", "error", err)
		os.Exit(1)
	}
	os.Exit(mainLoop(cflags))
}
