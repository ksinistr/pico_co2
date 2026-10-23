package types

import (
	"cmp"
	"slices"
	"time"
)

type MeasurementType string

const (
	MeasurementTypeTime        MeasurementType = "time"
	MeasurementTypeCO2         MeasurementType = "co2"
	MeasurementTypeTemperature MeasurementType = "temperature"
	MeasurementTypeHumidity    MeasurementType = "humidity"
)

type Measurement struct {
	Type           MeasurementType
	SensorID       string
	SensorPriority int
	Value          float64
	FirstReadingAt time.Time
	LastUpdateAt   time.Time
}

type Measurements []Measurement

func (m Measurements) ToMapByType() map[MeasurementType]Measurements {
	result := make(map[MeasurementType]Measurements)

	for _, measurement := range m {
		result[measurement.Type] = append(result[measurement.Type], measurement)
	}

	return result
}

func (m Measurements) OrderByPriority() Measurements {
	m = slices.Clone(m)
	slices.SortStableFunc(m, func(a, b Measurement) int {
		return cmp.Compare(a.SensorPriority, b.SensorPriority)
	})

	return m
}

func (m Measurements) First() (Measurement, bool) {
	if len(m) == 0 {
		return Measurement{}, false
	}

	return m[0], true
}

type MeasurementsByType map[MeasurementType]Measurements

func (m Measurements) Grouped() MeasurementsByType {
	result := MeasurementsByType(m.ToMapByType())
	for kind, values := range result {
		result[kind] = values.OrderByPriority()
	}

	return result
}
