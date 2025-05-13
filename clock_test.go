package main

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

func TestSkewClock(t *testing.T) {
	var st int64 = 18
	var inc int64 = 5
	clock := newSkewClock(st)
	t1 := clock.Now().Unix()
	clock.forward(time.Second * time.Duration(inc))
	t2 := clock.Now().Unix()
	if t1 != st {
		t.Errorf("wanted %d, but got %v", st, t1)
	}
	if t2 != st+inc {
		t.Errorf("wanted %d, but got %v", st+inc, t2)
	}
}

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
