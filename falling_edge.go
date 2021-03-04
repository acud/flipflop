// Copyright (c) 2021 acud

package flipflop

import (
	"io"
	"time"
)

type BufferedInput interface {
	Trigger()
	io.Closer
}

type detector struct {
	t         time.Duration
	worstCase time.Duration

	buf  chan struct{}
	out  chan struct{}
	quit chan struct{}
}

// bufferTime is the time to buffer, worstCase is buffertime*worstcase time to wait before writing
// to the output anyway.
func NewFallingEdge(bufferTime, worstCase time.Duration) (in chan<- struct{}, out <-chan struct{}, clean func()) {
	d := &detector{
		t:         bufferTime,
		worstCase: worstCase,
		buf:       make(chan struct{}, 1),
		out:       make(chan struct{}),
		quit:      make(chan struct{}),
	}

	go d.work()

	return d.buf, d.out, func() { close(d.quit) }
}

func (d *detector) work() {
	var waitWrite <-chan time.Time
	var worstCase <-chan time.Time
	for {
		select {
		case <-d.quit:
			return
		case <-d.buf:
			// we have an item in the buffer, dont announce yet
			waitWrite = time.After(d.t)
			if worstCase == nil {
				worstCase = time.After(d.worstCase)
			}
		case <-waitWrite:
			d.out <- struct{}{}
			worstCase = nil
			waitWrite = nil
		case <-worstCase:
			d.out <- struct{}{}
			worstCase = nil
			waitWrite = nil
		}

	}
}

// Triggers the input. Does not guarantee an item is put to the buffer.
func (d *detector) Trigger() {
	select {
	case d.buf <- struct{}{}:
	default:
	}
}

func (d *detector) Close() error {
	close(d.quit)
	return nil
}
