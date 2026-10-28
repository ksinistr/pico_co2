package sensors

import (
	"fmt"
	"math"
	"pico_co2/internal/types"
	"time"
)

type Schedule struct {
	Interval        time.Duration
	StartupInterval time.Duration
	StartupWindow   time.Duration
}

type readTimer struct {
	schedule  Schedule
	startedAt time.Time
	nextAt    time.Time
}

func newReadTimer(schedule Schedule) readTimer {
	if schedule.Interval <= 0 || schedule.StartupInterval < 0 || schedule.StartupWindow < 0 {
		panic("sensors: invalid read schedule")
	}

	return readTimer{schedule: schedule}
}

func (t *readTimer) due(now time.Time) bool {
	if now.Before(t.nextAt) {
		return false
	}

	if t.startedAt.IsZero() {
		t.startedAt = now
	}

	interval := t.schedule.Interval
	if t.schedule.StartupInterval > 0 && now.Sub(t.startedAt) < t.schedule.StartupWindow {
		interval = t.schedule.StartupInterval
	}

	t.nextAt = now.Add(interval)

	return true
}

func ready(reading types.Reading, value float64, now time.Time) types.Reading {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return failed(reading, fmt.Errorf("sensor returned invalid value: %v", value))
	}

	return types.Reading{Value: value, UpdatedAt: now, State: types.ReadingReady}
}

func failed(reading types.Reading, err error) types.Reading {
	reading.State = types.ReadingFailed
	reading.Err = err

	return reading
}

func warming(reading types.Reading) types.Reading {
	reading.State = types.ReadingWarmingUp
	reading.Err = nil

	return reading
}
