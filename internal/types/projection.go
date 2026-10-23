package types

import (
	"pico_co2/internal/types/status"
	"pico_co2/pkg/fifo"
	"time"
)

func (r *Readings) Project() {
	r.Raw = RawReadings{}

	r.Time = Time{}
	if m, ok := r.Measurements[MeasurementTypeCO2].First(); ok {
		r.Raw.CO2 = uint16(m.Value)
	}

	if m, ok := r.Measurements[MeasurementTypeTemperature].First(); ok {
		r.Raw.Temperature = float32(m.Value)
	}

	if m, ok := r.Measurements[MeasurementTypeHumidity].First(); ok {
		r.Raw.Humidity = float32(m.Value)
	}

	if m, ok := r.Measurements[MeasurementTypeTime].First(); ok {
		value := time.Unix(int64(m.Value), 0).UTC()
		r.Time = Time{Hour: value.Hour(), Minute: value.Minute(), LastRead: m.LastUpdateAt}
	}

	r.History.CO2 = r.selectedHistory(MeasurementTypeCO2)
	r.History.Temperature = r.selectedHistory(MeasurementTypeTemperature)
	r.History.Humidity = r.selectedHistory(MeasurementTypeHumidity)
	r.Calculated.CO2Trend = status.UnknownCO2Trend
	r.calculateCO2Trend()
}

func (r *Readings) selectedHistory(kind MeasurementType) *fifo.FIFO16 {
	value, ok := r.Measurements[kind].First()
	if ok {
		if history := r.SensorHistory[MeasurementKey{SensorID: value.SensorID, Type: kind}]; history != nil {
			return history
		}
	}

	return fifo.NewFIFO16(1)
}
