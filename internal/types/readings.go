package types

import (
	"math"
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
	LastRaw        RawReadings
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

// Clock edit field constants mirror clockedit.EditField values for use
// by the display layer without importing the clockedit package.
const (
	EditFieldHour   = 0
	EditFieldMinute = 1
	EditFieldSave   = 2
	EditFieldCancel = 3
)

// ClockEdit carries the display-relevant state of the clock editor.
type ClockEdit struct {
	Active bool
	Field  int // one of EditFieldHour, EditFieldMinute, EditFieldSave, EditFieldCancel
	Hour   int
	Minute int
}

type RawReadings struct {
	Temperature float32
	Humidity    float32
	CO2         uint16
	TVOC        uint16
	AQI         uint8
}

type MeasurementHistory struct {
	CO2           *fifo.FIFO16
	Temperature   *fifo.FIFO16
	Humidity      *fifo.FIFO16
	HeatIndexTemp *fifo.FIFO16
	DewPoint      *fifo.FIFO16
	AddedAt       time.Time
	Granularity   time.Duration
}

type CalculatedReadings struct {
	CO215MinAverage     uint16
	CO25MinAvgPrev      uint16
	CO25MinAvgCurr      uint16
	CO2Trend            status.CO2Trend
	DewPointC           float32
	AbsoluteHumidityGM3 float32
}

func InitReadings(queueSize int) *Readings {
	return &Readings{
		History: MeasurementHistory{
			CO2:           fifo.NewFIFO16(queueSize),
			Temperature:   fifo.NewFIFO16(queueSize),
			Humidity:      fifo.NewFIFO16(queueSize),
			HeatIndexTemp: fifo.NewFIFO16(queueSize),
			DewPoint:      fifo.NewFIFO16(queueSize),
			Granularity:   time.Minute,
		},
		Calculated: CalculatedReadings{
			CO2Trend: status.UnknownCO2Trend,
		},
	}
}

func (r *Readings) AddReadings(
	co2 uint16,
	temperature float32,
	humidity float32,
) {
	r.Error = ""
	r.LastUpdateAt = time.Now()

	if r.FirstReadingAt.IsZero() {
		r.FirstReadingAt = time.Now()
	}

	dewPoint := status.DewPointC(temperature, humidity)
	r.Calculated.DewPointC = dewPoint
	r.Calculated.AbsoluteHumidityGM3 = status.AbsoluteHumidityGM3(temperature, humidity)

	if r.History.CO2 == nil || r.History.Temperature == nil ||
		r.History.Humidity == nil || r.History.HeatIndexTemp == nil ||
		r.History.DewPoint == nil {
		return
	}

	if time.Since(r.History.AddedAt) > r.History.Granularity {
		if co2 > 0 {
			r.History.CO2.Enqueue(int16(co2))
		}
		r.History.Temperature.Enqueue(int16(math.Round(float64(temperature))))
		r.History.Humidity.Enqueue(int16(math.Round(float64(humidity))))
		hiVal := status.HeatIndexVal(temperature, humidity)
		r.History.HeatIndexTemp.Enqueue(int16(math.Round(float64(hiVal))))
		r.History.DewPoint.Enqueue(int16(math.Round(float64(dewPoint))))
		r.History.AddedAt = time.Now()
	}

	// Calculate 15-minute average of last 15 readings
	if r.History.CO2.Len() >= 15 {
		var sum uint32
		count := 0

		r.History.CO2.PeekAll(func(val int16) {
			sum += uint32(val)
			count++
			if count >= 15 {
				return
			}
		})

		if count > 0 {
			r.Calculated.CO215MinAverage = uint16(sum / uint32(count))
		}
	} else {
		// Not enough data yet, use current reading as initial average
		if r.Calculated.CO215MinAverage == 0 {
			r.Calculated.CO215MinAverage = co2
		}
	}

	// Calculate CO2 trend based on 5-minute moving averages
	r.calculateCO2Trend()

	// Store last measurements before updating with new ones
	r.LastRaw = r.Raw

	r.Raw = RawReadings{
		CO2:         co2,
		Temperature: temperature,
		Humidity:    humidity,
	}
}

func (r *Readings) calculateCO2Trend() {
	// Need at least 10 readings for two 5-minute windows
	if r.History.CO2.Len() < 10 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}

	readings := r.History.CO2.Contiguous()
	if len(readings) < 10 {
		r.Calculated.CO2Trend = status.UnknownCO2Trend
		return
	}

	// Previous 5-minute average (readings[-10:-5])
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

	// Current 5-minute average (readings[-5:])
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

	// Calculate trend
	diff := int32(currAvg) - int32(prevAvg)
	switch {
	case diff > 50:
		r.Calculated.CO2Trend = status.RisingCO2
	case diff < -50:
		r.Calculated.CO2Trend = status.FallingCO2
	default:
		r.Calculated.CO2Trend = status.StableCO2
	}

	// Store averages for debugging/analysis
	r.Calculated.CO25MinAvgPrev = prevAvg
	r.Calculated.CO25MinAvgCurr = currAvg
}
