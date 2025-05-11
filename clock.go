package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"
)

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type skewClock struct {
	skew time.Duration
}

func (cl *skewClock) Now() time.Time {
	return time.Now().Add(-cl.skew)
}

func newSkewClock(i int64) *skewClock {
	d := time.Now().Sub(time.Unix(i, 0))
	return &skewClock{skew: d}
}

func (cl *skewClock) forward(d time.Duration) {
	cl.skew -= d
}

// repeatUntilInterrupt runs the given function every interval.
// It will return whenever one of the signals in interruptSignals is received.
func repeatUntilInterrupt(callback func(), interval time.Duration, interruptSignals ...os.Signal) {
	slog.Info("running continuously")
	slog.Debug("loop timing", "interval", DefaultLoopInterval)
	ticker := time.NewTicker(interval)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, interruptSignals...)
	quit := make(chan bool)
	for {
		select {
		case <-ticker.C:
			callback()
		case sig := <-sigc:
			slog.Debug("received signal", "signal", sig)
			go func() { quit <- true }()
		case <-quit:
			ticker.Stop()
			slog.Debug("quit")
			return
		}
	}
}
