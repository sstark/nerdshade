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
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, interruptSignals...)
	quit := make(chan bool)
	acpiEvent, acpiErr := AcpiEvent(acpiListenCmd, acpiLidOpenEvent)
	if acpiErr != nil {
		slog.Warn("ACPI event listener could not be started", "error", acpiErr)
	}
	for {
		select {
		case <-ticker.C:
			callback()
		case <-acpiEvent:
			slog.Info("acpi event received")
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
