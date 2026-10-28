package sensors

import (
	"encoding/binary"
	"errors"
	"pico_co2/internal/types"
	"testing"
	"time"

	"tinygo.org/x/drivers/scd4x"
)

type scd4xBus struct {
	command   uint16
	available bool
	samples   int
	err       error
}

func (b *scd4xBus) Tx(_ uint16, write, read []byte) error {
	if b.err != nil {
		return b.err
	}
	if len(write) >= 2 {
		b.command = binary.BigEndian.Uint16(write)
	}
	if len(read) == 0 {
		return nil
	}
	clear(read)
	switch b.command {
	case scd4x.CmdDataReady:
		if b.available {
			read[1] = 1
		}
	case scd4x.CmdReadMeasurement:
		b.samples++
		b.available = false
		binary.BigEndian.PutUint16(read[0:2], 800)
		binary.BigEndian.PutUint16(read[3:5], 26214)
		binary.BigEndian.PutUint16(read[6:8], 32768)
	}
	return nil
}

func TestSCD4xDriverUsesOneSample(t *testing.T) {
	for _, tt := range []struct {
		name            string
		available, fail bool
		samples         int
		state           types.ReadingState
	}{
		{"one sample gives all quantities", true, false, 1, types.ReadingReady},
		{"not ready does not read sample", false, false, 0, types.ReadingWarmingUp},
		{"bus failure", false, true, 0, types.ReadingFailed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			bus := &scd4xBus{available: tt.available}
			if tt.fail {
				bus.err = errors.New("i2c disconnected")
			}
			source := NewSCD4xSource(scd4x.New(bus), Schedule{Interval: time.Second})
			got := source.Read(now)
			if bus.samples != tt.samples || got.CO2.State != tt.state {
				t.Fatalf("samples = %d, CO2 = %+v", bus.samples, got.CO2)
			}
			if tt.state != types.ReadingReady {
				return
			}
			if got.CO2.Value != 800 || got.Temperature.Value < 24.9 || got.Temperature.Value > 25.1 || got.Humidity.Value != 50 {
				t.Fatalf("inconsistent quantities: %+v", got)
			}
			cached := source.Read(now.Add(time.Second))
			if cached != got || bus.samples != 1 {
				t.Fatal("no new sample changed values or timestamp")
			}
		})
	}
}
