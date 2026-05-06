package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestEventLoop(t *testing.T) {
	ticks := make(chan time.Time)
	sigc := make(chan os.Signal, 1)
	acpiEvents := make(chan string, 10)
	called := make(chan struct{}, 10)
	done := make(chan struct{})

	go func() {
		eventLoop(func() { called <- struct{}{} }, ticks, sigc, acpiEvents)
		close(done)
	}()

	expectCallback := func(reason string) {
		t.Helper()
		select {
		case <-called:
		case <-time.After(5 * time.Second):
			t.Fatalf("callback was not called after %s", reason)
		}
	}

	ticks <- time.Now()
	expectCallback("tick")

	acpiEvents <- acpiLidOpenEvent
	expectCallback("acpi event")

	sigc <- syscall.SIGINT
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("eventLoop did not return after a signal")
	}

	select {
	case <-called:
		t.Error("callback was called for the signal")
	default:
	}
}
