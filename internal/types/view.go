package types

import (
	"pico_co2/internal/clockedit"
	"pico_co2/internal/types/status"
	"time"
)

type Time struct {
	Hour     int
	Minute   int
	LastRead time.Time
}

type ClockEdit struct {
	Active bool
	Field  clockedit.EditField
	Hour   int
	Minute int
}

type ViewState struct {
	Readings  Readings
	History   *History
	CO2Trend  status.CO2Trend
	Time      Time
	ClockEdit ClockEdit
}
