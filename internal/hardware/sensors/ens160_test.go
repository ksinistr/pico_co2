package sensors

import (
	"errors"
	"math"
	"pico_co2/internal/types"
	"testing"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ens160"
)

type fakeAHT20 struct {
	err         error
	temperature float32
	reads       int
}

func (d *fakeAHT20) Read() error          { d.reads++; return d.err }
func (d *fakeAHT20) Celsius() float32     { return d.temperature }
func (d *fakeAHT20) RelHumidity() float32 { return 55 }

type fakeENS160 struct {
	err             error
	compensationErr error
	validity        uint8
	reads           int
	compensations   int
	temperature     int32
	humidity        int32
}

func (d *fakeENS160) SetEnvDataMilli(temperature, humidity int32) error {
	d.compensations++
	d.temperature, d.humidity = temperature, humidity
	return d.compensationErr
}
func (d *fakeENS160) Update(which drivers.Measurement) error { d.reads++; return d.err }
func (d *fakeENS160) ECO2() uint16                           { return 900 }
func (d *fakeENS160) Validity() uint8                        { return d.validity }

func TestSeparateAHT20AndENS160Sources(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name               string
		climateFail        bool
		airFail            bool
		compensationFail   bool
		invalidTemperature bool
		validity           uint8
		climateState       types.ReadingState
		airState           types.ReadingState
	}{
		{"ready", false, false, false, false, ens160.ValidityNormalOperation, types.ReadingReady, types.ReadingReady},
		{"air warmup", false, false, false, false, ens160.ValidityWarmUpPhase, types.ReadingReady, types.ReadingWarmingUp},
		{"air initial startup", false, false, false, false, ens160.ValidityInitialStartUpPhase, types.ReadingReady, types.ReadingWarmingUp},
		{"invalid air output", false, false, false, false, ens160.ValidityInvalidOutput, types.ReadingReady, types.ReadingFailed},
		{"climate fails independently", true, false, false, false, ens160.ValidityNormalOperation, types.ReadingFailed, types.ReadingReady},
		{"air fails independently", false, true, false, false, ens160.ValidityNormalOperation, types.ReadingReady, types.ReadingFailed},
		{"compensation fails", false, false, true, false, ens160.ValidityNormalOperation, types.ReadingReady, types.ReadingFailed},
		{"nonfinite temperature", false, false, false, true, ens160.ValidityNormalOperation, types.ReadingFailed, types.ReadingReady},
	} {
		t.Run(tt.name, func(t *testing.T) {
			climate := &fakeAHT20{temperature: 24.5}
			air := &fakeENS160{validity: tt.validity}
			if tt.climateFail {
				climate.err = errors.New("climate disconnected")
			}
			if tt.airFail {
				air.err = errors.New("air disconnected")
			}
			if tt.compensationFail {
				air.compensationErr = errors.New("write failed")
			}
			if tt.invalidTemperature {
				climate.temperature = float32(math.NaN())
			}
			schedule := Schedule{Interval: time.Second}
			climateSource := NewAHT20Source(climate, schedule)
			source := NewENS160Source(air, climateSource, schedule)
			got := climateSource.Read(now)
			got.CO2 = source.Read(now).CO2
			if got.Temperature.State != tt.climateState || got.CO2.State != tt.airState {
				t.Fatalf("readings = %+v", got)
			}
			if tt.airState == types.ReadingReady && got.CO2.Value != 900 {
				t.Fatal("wrong CO2")
			}
			if tt.airState != types.ReadingReady && !got.CO2.UpdatedAt.IsZero() {
				t.Fatal("nonready CO2 has timestamp")
			}
			if !tt.climateFail && !tt.invalidTemperature && (air.temperature != 24500 || air.humidity != 55000) {
				t.Fatalf("compensation units: %d %d", air.temperature, air.humidity)
			}
			if (tt.climateFail || tt.invalidTemperature) && air.compensations != 0 {
				t.Fatal("used invalid compensation")
			}
			source.Read(now.Add(500 * time.Millisecond))
			if climate.reads != 1 || air.reads > 1 {
				t.Fatal("cached read touched hardware")
			}
		})
	}
}

func TestSeparateSensorIntervals(t *testing.T) {
	for _, tt := range []struct {
		name                         string
		climateInterval, airInterval time.Duration
		climateReads, airReads       int
	}{
		{"climate slower", time.Minute, time.Second, 1, 2},
		{"air slower", time.Second, time.Minute, 2, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			climate := &fakeAHT20{temperature: 24.5}
			air := &fakeENS160{}
			climateSource := NewAHT20Source(climate, Schedule{Interval: tt.climateInterval})
			source := NewENS160Source(air, climateSource, Schedule{Interval: tt.airInterval})
			climateSource.Read(now)
			source.Read(now)
			climateSource.Read(now.Add(time.Second))
			source.Read(now.Add(time.Second))
			if climate.reads != tt.climateReads || air.reads != tt.airReads {
				t.Fatal("intervals were coupled")
			}
		})
	}
}
