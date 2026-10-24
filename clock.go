package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"
)

const (
	acpiLidOpenEvent = "button/lid LID open"
	acpiListenCmd    = "acpi_listen"
)

// repeatUntilInterrupt runs the given function every interval.
// It will return whenever one of the signals in interruptSignals is received.
func repeatUntilInterrupt(callback func(), interval time.Duration, interruptSignals ...os.Signal) {
	slog.Info("running continuously")
	slog.Debug("loop timing", "interval", DefaultLoopInterval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, interruptSignals...)
	defer signal.Stop(sigc)
	acpiEvent, acpiStop, acpiErr := AcpiEvent(acpiListenCmd, acpiLidOpenEvent)
	if acpiStop != nil {
		defer acpiStop()
	}
	if acpiErr != nil {
		slog.Warn("ACPI event listener could not be started", "error", acpiErr)
	}
	eventLoop(callback, ticker.C, sigc, acpiEvent)
}

// eventLoop calls callback on every tick in ticks and on every event in
// acpiEvents and returns as soon as a signal arrives on sigc.
func eventLoop(callback func(), ticks <-chan time.Time, sigc <-chan os.Signal, acpiEvents <-chan string) {
	for {
		select {
		case <-ticks:
			callback()
		case <-acpiEvents:
			slog.Info("acpi event received")
			callback()
		case sig := <-sigc:
			slog.Debug("received signal", "signal", sig)
			return
		}
	}
}
