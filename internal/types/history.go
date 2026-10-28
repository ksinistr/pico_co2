package types

import (
	"math"
	"pico_co2/internal/types/status"
	"pico_co2/pkg/fifo"
	"time"
)

type Series struct {
	Values      *fifo.FIFO16
	LastAddedAt time.Time
}

type History struct {
	Temperature Series
	Humidity    Series
	CO2         Series
	Interval    time.Duration
}

func NewHistory(capacity int, interval time.Duration) *History {
	if capacity <= 0 || interval <= 0 {
		panic("types: history capacity and interval must be positive")
	}

	return &History{
		Temperature: Series{Values: fifo.NewFIFO16(capacity)},
		Humidity:    Series{Values: fifo.NewFIFO16(capacity)},
		CO2:         Series{Values: fifo.NewFIFO16(capacity)},
		Interval:    interval,
	}
}

func (h *History) Add(readings Readings) {
	h.Temperature.add(readings.Temperature, h.Interval)
	h.Humidity.add(readings.Humidity, h.Interval)
	h.CO2.add(readings.CO2, h.Interval)
}

func (s *Series) add(reading Reading, interval time.Duration) {
	if s.Values == nil || reading.State != ReadingReady || reading.UpdatedAt.IsZero() {
		return
	}

	if !s.LastAddedAt.IsZero() && reading.UpdatedAt.Sub(s.LastAddedAt) < interval {
		return
	}

	value := math.Round(reading.Value)
	if math.IsNaN(value) || value < math.MinInt16 || value > math.MaxInt16 {
		return
	}

	s.Values.Enqueue(int16(value))
	s.LastAddedAt = reading.UpdatedAt
}

func (h *History) CO2Trend() status.CO2Trend {
	if h.CO2.Values == nil {
		return status.UnknownCO2Trend
	}

	readings := h.CO2.Values.Contiguous()
	if len(readings) < 10 {
		return status.UnknownCO2Trend
	}

	previous := averageCO2(readings[len(readings)-10 : len(readings)-5])
	current := averageCO2(readings[len(readings)-5:])
	difference := current - previous

	switch {
	case difference > 50:
		return status.RisingCO2
	case difference < -50:
		return status.FallingCO2
	default:
		return status.StableCO2
	}
}

func averageCO2(readings []int16) int32 {
	var sum int32
	for _, reading := range readings {
		sum += int32(reading)
	}

	return sum / int32(len(readings))
}
