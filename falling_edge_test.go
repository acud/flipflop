// Copyright (c) 2021 acud

package flipflop_test

import (
	"testing"
	"time"

	flipflop "github.com/acud/flipflop"
)

func TestFallingEdge(t *testing.T) {
	ok := make(chan struct{})
	tt := 50 * time.Millisecond
	worst := 5 * tt
	in, c, cleanup := flipflop.NewFallingEdge(tt, worst)
	defer cleanup()
	go func() {
		select {
		case <-c:
			close(ok)
			return
		case <-time.After(100 * time.Millisecond):
			t.Errorf("timed out")
		}
	}()

	in <- struct{}{}

	select {
	case <-ok:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out")
	}
}

func TestFallingEdgeBuffer(t *testing.T) {
	ok := make(chan struct{})
	tt := 100 * time.Millisecond
	worst := 9 * tt
	in, c, cleanup := flipflop.NewFallingEdge(tt, worst)
	defer cleanup()
	sleeps := 5
	wait := 99 * time.Millisecond

	start := time.Now()

	go func() {
		select {
		case <-c:
			if time.Since(start) <= 450*time.Millisecond {
				t.Errorf("wrote too early %v", time.Since(start))
			}
			close(ok)
			return
		case <-time.After(1000 * time.Millisecond):
			t.Errorf("timed out")
		}
	}()
	for i := 0; i < sleeps; i++ {
		in <- struct{}{}
		time.Sleep(wait)
	}
	select {
	case <-ok:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out")
	}
}

func TestFallingEdgeWorstCase(t *testing.T) {
	ok := make(chan struct{})
	tt := 100 * time.Millisecond
	worst := 5 * tt
	in, c, cleanup := flipflop.NewFallingEdge(tt, worst)
	defer cleanup()
	sleeps := 9
	wait := 80 * time.Millisecond

	start := time.Now()

	go func() {
		select {
		case <-c:
			if time.Since(start) >= 550*time.Millisecond {
				t.Errorf("wrote too early %v", time.Since(start))
			}

			close(ok)
			return
		case <-time.After(1000 * time.Millisecond):
			t.Errorf("timed out")
		}
	}()
	go func() {
		for i := 0; i < sleeps; i++ {
			in <- struct{}{}
			time.Sleep(wait)
		}
	}()
	select {
	case <-ok:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out")
	}
}
