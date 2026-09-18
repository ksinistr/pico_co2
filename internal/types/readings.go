package types

import (
	"math"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/types/status"
	"pico_co2/pkg/fifo"
	"time"
)

type Readings struct {
	Raw            RawReadings
	Calculated     CalculatedReadings
	History        MeasurementHistory
	FirstReadingAt time.Time
	LastUpdateAt   time.Time
	IsDrawen       bool
	Error          string
	Time           Time
	ClockEdit      ClockEdit
}

type Time struct {
	Hour     int
	Minute   int
	LastRead time.Time
}

const DefaultHistoryCapacity = 480

// ClockEdit carries the display-relevant state of the clock editor.
type ClockEdit struct {
	Active bool
	Field  clockedit.EditField
	Hour   int
	Minute int
}

type RawReadings struct {
	Temperature float32
	Humidity    float32
	CO2         uint16
}

type MeasurementHistory struct {
	CO2         *fifo.FIFO16
	Temperature *fifo.FIFO16
	Humidity    *fifo.FIFO16
	AddedAt     time.Time
	Granularity time.Duration
}

type CalculatedReadings struct {
	CO2Trend status.CO2Trend
}

func InitReadings(queueSize int, historyInterval time.Duration) *Readings {
	return &Readings{
		History: MeasurementHistory{
			CO2:         fifo.NewFIFO16(queueSize),
			Temperature: fifo.NewFIFO16(queueSize),
			Humidity:    fifo.NewFIFO16(queueSize),
			Granularity: historyInterval,
		},
		Calculated: CalculatedReadings{
			CO2Trend: status.UnknownCO2Trend,
		},
	}
}

func (r *Readings) AddReadingsAt(
	now time.Time,
	co2 uint16,
	temperature float32,
	humidity float32,
) {
	r.Error = ""
	r.LastUpdateAt = now

	if r.FirstReadingAt.IsZero() {
		r.FirstReadingAt = now
	}

	if r.History.CO2 == nil || r.History.Temperature == nil ||
		r.History.Humidity == nil {
		return
	}

	if now.Sub(r.History.AddedAt) >= r.History.Granularity {
		if co2 > 0 {
			r.History.CO2.Enqueue(int16(co2))
		}

		r.History.Temperature.Enqueue(int16(math.Round(float64(temperature))))
		r.History.Humidity.Enqueue(int16(math.Round(float64(humidity))))
		r.History.AddedAt = now
	}

	r.calculateCO2Trend()

	r.Raw = RawReadings{
		CO2:         co2,
		Temperature: temperature,
		Humidity:    humidity,
	}
}

func (r *Readings) calculateCO2Trend() {
	readings := r.History.CO2.Contiguous()
	if len(readings) < 10 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}

	prevAvg := averageCO2(readings[len(readings)-10 : len(readings)-5])
	currAvg := averageCO2(readings[len(readings)-5:])

	diff := int32(currAvg) - int32(prevAvg)
	switch {
	case diff > 50:
		r.Calculated.CO2Trend = status.RisingCO2
	case diff < -50:
		r.Calculated.CO2Trend = status.FallingCO2
	default:
		r.Calculated.CO2Trend = status.StableCO2
	}
}

func averageCO2(readings []int16) uint16 {
	var sum uint32
	for _, reading := range readings {
		sum += uint32(reading)
	}

	return uint16(sum / uint32(len(readings)))
}
