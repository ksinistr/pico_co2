package sensors

import (
	"errors"
	"pico_co2/internal/config"
	"pico_co2/internal/types"
	"testing"
	"time"
)

type aht20Bus struct {
	err     error
	reads   int
	address uint16
}

func (b *aht20Bus) Tx(address uint16, _ []byte, read []byte) error {
	b.address = address
	if b.err != nil {
		return b.err
	}
	if len(read) == 1 {
		read[0] = 0x08
	}
	if len(read) == 7 {
		b.reads++
		copy(read, []byte{0x08, 0x80, 0, 0x06, 0, 0, 0})
	}
	return nil
}

func TestAHT20ConstructionAndRead(t *testing.T) {
	for _, tt := range []struct {
		name    string
		fail    bool
		address uint16
	}{
		{name: "default address", address: 0x38},
		{name: "custom address", address: 0x39},
		{name: "bus error", address: 0x38, fail: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bus := &aht20Bus{}
			if tt.fail {
				bus.err = errors.New("disconnected")
			}
			address := tt.address
			if address == 0x38 {
				address = 0
			}
			source, err := New(config.Sensor{Model: config.AHT20, Address: address}, bus, nil, Schedule{Interval: time.Second})
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			got := source.Read(now)
			cached := source.Read(now.Add(50 * time.Millisecond))
			if cached != got || bus.address != tt.address {
				t.Fatal("cached result or address changed")
			}
			if tt.fail {
				if got.Temperature.State != types.ReadingFailed || got.Temperature.Err == nil || !got.Temperature.UpdatedAt.IsZero() {
					t.Fatal("missing read failure")
				}
				return
			}
			if got.Temperature.Value != 25 || got.Humidity.Value != 50 || bus.reads != 1 {
				t.Fatalf("readings = %+v, reads = %d", got, bus.reads)
			}
			if got.CO2.State != types.ReadingUnavailable || got.Temperature.UpdatedAt != now {
				t.Fatal("wrong reading availability or time")
			}
		})
	}
}
