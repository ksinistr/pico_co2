package app

import (
	"fmt"
	"math"
	"pico_co2/internal/sensors"
	"pico_co2/internal/types"
	"pico_co2/pkg/fifo"
	"slices"
	"time"
)

type Registration struct {
	ID              string
	Priority        int
	Capabilities    []types.MeasurementType
	Reader          sensors.Reader
	Interval        time.Duration
	StartupInterval time.Duration
	StartupWindow   time.Duration
}

type sensorState struct {
	registration   Registration
	values         types.Measurements
	attemptedAt    time.Time
	firstReadingAt time.Time
	err            error
}

func (s *sensorState) read(now time.Time) (bool, error) {
	registration := s.registration
	interval := registration.Interval

	starting := s.firstReadingAt.IsZero() || now.Sub(s.firstReadingAt) < registration.StartupWindow
	if registration.StartupInterval > 0 && starting {
		interval = registration.StartupInterval
	}

	if !s.attemptedAt.IsZero() && now.Sub(s.attemptedAt) < interval {
		return false, s.err
	}

	s.attemptedAt = now

	values, err := registration.Reader.Read()
	if err == nil {
		err = validateMeasurements(values, registration.Capabilities)
	}

	s.err = err
	if err != nil {
		return false, err
	}

	if s.firstReadingAt.IsZero() {
		s.firstReadingAt = now
	}

	s.values = slices.Clone(values)
	for i := range s.values {
		value := &s.values[i]
		value.SensorID = registration.ID
		value.SensorPriority = registration.Priority
		value.FirstReadingAt = s.firstReadingAt
		value.LastUpdateAt = now
	}

	return true, nil
}

type Collector struct {
	states          []sensorState
	History         map[types.MeasurementKey]*fifo.FIFO16
	historyAt       map[types.MeasurementKey]time.Time
	capacity        int
	historyInterval time.Duration
}

func NewCollector(registrations []Registration, capacity int, historyInterval time.Duration) *Collector {
	if capacity <= 0 || historyInterval <= 0 {
		panic("app: history capacity and interval must be positive")
	}

	c := &Collector{
		capacity:        capacity,
		historyInterval: historyInterval,
		History:         make(map[types.MeasurementKey]*fifo.FIFO16),
		historyAt:       make(map[types.MeasurementKey]time.Time),
	}

	ids := make(map[string]bool)
	for _, registration := range registrations {
		validateRegistration(registration, ids)
		ids[registration.ID] = true
		registration.Capabilities = slices.Clone(registration.Capabilities)
		c.states = append(c.states, sensorState{registration: registration})
	}

	return c
}

func validateRegistration(registration Registration, ids map[string]bool) {
	if registration.ID == "" || ids[registration.ID] || registration.Reader == nil {
		panic("app: sensor needs a unique ID and reader")
	}

	if registration.Interval <= 0 || registration.StartupInterval < 0 || registration.StartupWindow < 0 {
		panic("app: invalid sensor interval")
	}

	if len(registration.Capabilities) == 0 {
		panic("app: sensor needs measurement capabilities")
	}
}

func (c *Collector) Invalidate(kind types.MeasurementType) {
	for i := range c.states {
		if slices.Contains(c.states[i].registration.Capabilities, kind) {
			c.states[i].attemptedAt = time.Time{}
		}
	}
}

func (c *Collector) Read(now time.Time) (types.Measurements, error) {
	var result types.Measurements

	for i := range c.states {
		state := &c.states[i]

		changed, err := state.read(now)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", state.registration.ID, err)
		}

		if changed {
			c.recordHistory(state.values, now)
		}

		result = append(result, state.values...)
	}

	return result, nil
}

func (c *Collector) recordHistory(values types.Measurements, now time.Time) {
	for _, value := range values {
		if value.Type == types.MeasurementTypeTime || (value.Type == types.MeasurementTypeCO2 && value.Value <= 0) {
			continue
		}

		key := types.MeasurementKey{SensorID: value.SensorID, Type: value.Type}
		if c.History[key] == nil {
			c.History[key] = fifo.NewFIFO16(c.capacity)
		}

		if c.historyAt[key].IsZero() || now.Sub(c.historyAt[key]) >= c.historyInterval {
			c.History[key].Enqueue(int16(math.Round(value.Value)))
			c.historyAt[key] = now
		}
	}
}

func validateMeasurements(values types.Measurements, capabilities []types.MeasurementType) error {
	seen := make(map[types.MeasurementType]bool)
	for _, value := range values {
		if !slices.Contains(capabilities, value.Type) || seen[value.Type] {
			return fmt.Errorf("unexpected measurement: %s", value.Type)
		}

		if math.IsNaN(value.Value) || math.IsInf(value.Value, 0) {
			return fmt.Errorf("invalid measurement: %s", value.Type)
		}

		seen[value.Type] = true
	}

	for _, kind := range capabilities {
		if !seen[kind] {
			return fmt.Errorf("missing measurement: %s", kind)
		}
	}

	return nil
}
