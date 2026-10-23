package types

import (
	"pico_co2/pkg/fifo"
	"testing"
	"time"
)

func TestGroupedMeasurements(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values Measurements
		want   string
		exists bool
	}{
		{"empty", nil, "", false},
		{"priority", Measurements{{Type: MeasurementTypeTemperature, SensorID: "a", SensorPriority: 2}, {Type: MeasurementTypeTemperature, SensorID: "b", SensorPriority: 1}}, "b", true},
		{"stable ties", Measurements{{Type: MeasurementTypeTemperature, SensorID: "a", SensorPriority: 1}, {Type: MeasurementTypeTemperature, SensorID: "b", SensorPriority: 1}}, "a", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			grouped := tt.values.Grouped()
			selected, ok := grouped[MeasurementTypeTemperature].First()
			if ok != tt.exists || selected.SensorID != tt.want {
				t.Fatalf("selected=%+v,ok=%v", selected, ok)
			}
			if len(tt.values) > 0 && tt.values[0].SensorID != "a" {
				t.Fatal("grouping modified input")
			}
		})
	}
}

func TestProjection(t *testing.T) {
	for _, tt := range []struct {
		name     string
		priority int
		want     string
	}{
		{"first source", 0, "first"}, {"second source", 2, "second"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 10, 21, 14, 23, 0, 0, time.UTC)
			r := InitReadings(4, time.Minute)
			r.Measurements = Measurements{
				{Type: MeasurementTypeTemperature, SensorID: "first", SensorPriority: tt.priority, Value: 21},
				{Type: MeasurementTypeTemperature, SensorID: "second", SensorPriority: 1, Value: 25},
				{Type: MeasurementTypeTime, SensorID: "clock", Value: float64(now.Unix())},
			}.Grouped()
			first, second := fifo.NewFIFO16(4), fifo.NewFIFO16(4)
			first.Enqueue(21)
			second.Enqueue(25)
			r.SensorHistory = map[MeasurementKey]*fifo.FIFO16{
				{SensorID: "first", Type: MeasurementTypeTemperature}:  first,
				{SensorID: "second", Type: MeasurementTypeTemperature}: second,
			}
			r.Project()
			selected, _ := r.Measurements[MeasurementTypeTemperature].First()
			if selected.SensorID != tt.want || r.Raw.Temperature != float32(selected.Value) {
				t.Fatalf("projection=%+v", r.Raw)
			}
			if r.History.Temperature != r.SensorHistory[MeasurementKey{SensorID: tt.want, Type: MeasurementTypeTemperature}] {
				t.Fatal("mixed source histories")
			}
			if r.Time.Hour != 14 || r.Time.Minute != 23 {
				t.Fatalf("time=%+v", r.Time)
			}
		})
	}
}
