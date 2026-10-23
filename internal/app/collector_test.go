package app

import (
	"errors"
	"pico_co2/internal/types"
	"strings"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	for _, tt := range []struct {
		name string
		fail bool
	}{{"independent schedules and histories", false}, {"stop on error and recover", true}} {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			firstReads, secondReads := 0, 0
			failing := tt.fail
			c := NewCollector([]Registration{
				{ID: "first", Priority: 2, Capabilities: []types.MeasurementType{types.MeasurementTypeTemperature}, Interval: time.Second, Reader: readerFunc(func() ([]types.Measurement, error) {
					firstReads++
					if failing {
						return nil, errors.New("broken")
					}
					return []types.Measurement{{Type: types.MeasurementTypeTemperature, Value: 20.4}}, nil
				})},
				{ID: "second", Priority: 1, Capabilities: []types.MeasurementType{types.MeasurementTypeTemperature}, Interval: time.Minute, Reader: readerFunc(func() ([]types.Measurement, error) {
					secondReads++
					return []types.Measurement{{Type: types.MeasurementTypeTemperature, Value: 22.6}}, nil
				})},
			}, 4, time.Minute)
			values, err := c.Read(now)
			if tt.fail {
				if err == nil || !strings.Contains(err.Error(), "first: broken") || secondReads != 0 {
					t.Fatalf("error=%v second reads=%d", err, secondReads)
				}
				failing = false
				if _, err = c.Read(now.Add(time.Millisecond)); err == nil {
					t.Fatal("error cleared before retry")
				}
				values, err = c.Read(now.Add(time.Second))
			}
			if err != nil {
				t.Fatal(err)
			}
			selected, ok := values.Grouped()[types.MeasurementTypeTemperature].First()
			if !ok || selected.SensorID != "second" || selected.Value != 22.6 {
				t.Fatalf("selected=%+v", selected)
			}
			if _, err = c.Read(now.Add(2 * time.Second)); err != nil {
				t.Fatal(err)
			}
			if secondReads != 1 || firstReads < 2 {
				t.Fatalf("reads=%d,%d", firstReads, secondReads)
			}
			for _, id := range []string{"first", "second"} {
				history := c.History[types.MeasurementKey{SensorID: id, Type: types.MeasurementTypeTemperature}]
				if history == nil || history.Len() != 1 {
					t.Fatalf("history for %s=%v", id, history)
				}
			}
		})
	}
}

func TestRegistrationValidation(t *testing.T) {
	for _, tt := range []struct {
		name      string
		values    types.Measurements
		wantError bool
	}{
		{"valid", types.Measurements{{Type: types.MeasurementTypeCO2, Value: 700}}, false},
		{"missing", nil, true},
		{"undeclared", types.Measurements{{Type: types.MeasurementTypeHumidity}}, true},
		{"duplicate", types.Measurements{{Type: types.MeasurementTypeCO2}, {Type: types.MeasurementTypeCO2}}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMeasurements(tt.values, []types.MeasurementType{types.MeasurementTypeCO2})
			if (err != nil) != tt.wantError {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
