package main

import (
	"machine"
	"sync/atomic"
	"time"
)

type TouchButton struct {
	pin  machine.Pin
	flag uint32
	last time.Time
}

func NewTouchButton(p machine.Pin) *TouchButton {
	b := &TouchButton{pin: p}
	b.pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	b.pin.SetInterrupt(machine.PinRising, func(machine.Pin) {
		now := time.Now()
		if now.Sub(b.last) < 50*time.Millisecond {
			return
		}
		b.last = now
		atomic.StoreUint32(&b.flag, 1)
	})

	return b
}

func (b *TouchButton) Consume() bool {
	if atomic.SwapUint32(&b.flag, 0) != 0 {
		return true
	}

	return false
}
