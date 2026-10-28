package types

import "time"

type ReadingState uint8

const (
	ReadingUnavailable ReadingState = iota
	ReadingWarmingUp
	ReadingReady
	ReadingFailed
)

type Reading struct {
	Value     float64
	UpdatedAt time.Time
	State     ReadingState
	Err       error
}

type Readings struct {
	Temperature Reading
	Humidity    Reading
	CO2         Reading
}

const DefaultHistoryCapacity = 480
