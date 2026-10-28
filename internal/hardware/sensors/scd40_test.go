package sensors

import (
	"errors"
	"pico_co2/internal/types"
	"testing"
	"time"
)

type fakeSCD4x struct {
	available bool
	err       error
	reads     int
}

func (d *fakeSCD4x) DataReady() (bool, error)        { return d.available, d.err }
func (d *fakeSCD4x) ReadCO2() (int32, error)         { d.reads++; return 800, d.err }
func (d *fakeSCD4x) ReadTemperature() (int32, error) { return 24500, d.err }
func (d *fakeSCD4x) ReadHumidity() (int32, error)    { return 55, d.err }

func TestSCD4xSource(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name      string
		available bool
		fail      bool
		state     types.ReadingState
		reads     int
	}{
		{"ready", true, false, types.ReadingReady, 1},
		{"warming up", false, false, types.ReadingWarmingUp, 0},
		{"unreachable", false, true, types.ReadingFailed, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			device := &fakeSCD4x{available: tt.available}
			if tt.fail {
				device.err = errors.New("disconnected")
			}
			source := NewSCD4xSource(device, Schedule{Interval: time.Second})
			got := source.Read(now)
			for _, value := range []types.Reading{got.CO2, got.Temperature, got.Humidity} {
				if value.State != tt.state || (value.Err != nil) != tt.fail {
					t.Fatalf("reading = %+v", value)
				}
				if (value.State == types.ReadingReady) != value.UpdatedAt.Equal(now) {
					t.Fatalf("unexpected timestamp: %v", value.UpdatedAt)
				}
			}
			if device.reads != tt.reads {
				t.Fatalf("reads = %d", device.reads)
			}
			if tt.state == types.ReadingReady && (got.Temperature.Value != 24.5 || got.Humidity.Value != 55 || got.CO2.Value != 800) {
				t.Fatalf("units = %+v", got)
			}
		})
	}
}

func TestSCD4xScheduleAndRecovery(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name     string
		fail     bool
		notReady bool
	}{
		{"cached", false, false},
		{"not ready preserves timestamp", false, true},
		{"failure and recovery", true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			device := &fakeSCD4x{available: true}
			source := NewSCD4xSource(device, Schedule{Interval: time.Minute, StartupInterval: time.Second, StartupWindow: time.Minute})
			first := source.Read(now)
			cached := source.Read(now.Add(500 * time.Millisecond))
			if cached != first || device.reads != 1 {
				t.Fatal("cached read touched device")
			}
			if tt.fail {
				device.err = errors.New("disconnected")
			}
			device.available = !tt.notReady
			got := source.Read(now.Add(time.Second))
			if tt.fail || tt.notReady {
				if got.CO2.UpdatedAt != first.CO2.UpdatedAt || got.CO2.Value != first.CO2.Value {
					t.Fatal("unsuccessful read changed last successful value")
				}
			}
			device.err, device.available = nil, true
			got = source.Read(now.Add(time.Minute))
			if got.CO2.State != types.ReadingReady || got.CO2.Err != nil {
				t.Fatal("did not recover")
			}
			reads := device.reads
			source.Read(now.Add(time.Minute + time.Second))
			if device.reads != reads {
				t.Fatal("regular interval was not applied")
			}
		})
	}
}
