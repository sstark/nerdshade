package main

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

// On a very busy machine this test can be flaky
func TestRepeatUntilInterrupt(t *testing.T) {
	tests := []struct {
		totalRuntime int
		interval     int
	}{
		// For intervals < 10, notable delays happen too often on my machine
		{80, 30},
		{90, 40},
		{70, 300}, // callback is never called
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d/%d", test.totalRuntime, test.interval), func(t *testing.T) {
			var called []string
			// Kill ourselves after totalRuntime
			time.AfterFunc(time.Duration(test.totalRuntime)*time.Millisecond, func() {
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			})

			repeatUntilInterrupt(func() {
				called = append(called, "called")
			}, time.Duration(test.interval)*time.Millisecond, syscall.SIGINT)

			expectedCalls := test.totalRuntime / test.interval
			if len(called) != expectedCalls {
				t.Errorf("callback was not called the correct number of times. Expected %v, got %v", expectedCalls, len(called))
			}

		})
	}
}
