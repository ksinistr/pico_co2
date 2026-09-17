package types

import (
	"math"
	"time"

	"pico_co2/internal/clockedit"
	"pico_co2/internal/types/status"
	"pico_co2/pkg/fifo"
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
	if r.History.CO2.Len() < 10 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}

	readings := r.History.CO2.Contiguous()
	if len(readings) < 10 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}

	var prevSum uint32
	prevCount := 0
	for i := len(readings) - 10; i < len(readings)-5 && i >= 0; i++ {
		if i >= 0 && i < len(readings) {
			prevSum += uint32(readings[i])
			prevCount++
		}
	}

	if prevCount == 0 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}
	prevAvg := uint16(prevSum / uint32(prevCount))

	var currSum uint32
	currCount := 0
	for i := len(readings) - 5; i < len(readings) && i >= 0; i++ {
		if i >= 0 && i < len(readings) {
			currSum += uint32(readings[i])
			currCount++
		}
	}

	if currCount == 0 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}
	currAvg := uint16(currSum / uint32(currCount))

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
