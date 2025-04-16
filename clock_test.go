package main

import (
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
