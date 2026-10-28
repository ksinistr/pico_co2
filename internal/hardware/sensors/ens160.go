package sensors

import (
	"fmt"
	"math"
	"pico_co2/internal/config"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ens160"
)

type ENS160Device interface {
	SetEnvDataMilli(int32, int32) error
	Update(drivers.Measurement) error
	ECO2() uint16
	Validity() uint8
}

type CompensationReader interface {
	Read(time.Time) types.Readings
}

type ENS160Source struct {
	device       ENS160Device
	compensation CompensationReader
	reading      types.Reading
	timer        readTimer
}

func NewENS160Source(device ENS160Device, compensation CompensationReader, schedule Schedule) *ENS160Source {
	return &ENS160Source{device: device, compensation: compensation, timer: newReadTimer(schedule)}
}

func (s *ENS160Source) Read(now time.Time) types.Readings {
	if s.timer.due(now) {
		s.update(now)
	}

	return types.Readings{CO2: s.reading}
}

func (s *ENS160Source) update(now time.Time) {
	if err := s.compensate(now); err != nil {
		s.reading = failed(s.reading, err)
		return
	}

	if err := s.device.Update(drivers.Concentration); err != nil {
		s.reading = failed(s.reading, fmt.Errorf("ens160 read: %w", err))
		return
	}

	switch s.device.Validity() {
	case ens160.ValidityNormalOperation:
		s.reading = ready(s.reading, float64(s.device.ECO2()), now)
	case ens160.ValidityWarmUpPhase, ens160.ValidityInitialStartUpPhase:
		s.reading = warming(s.reading)
	default:
		s.reading = failed(s.reading, fmt.Errorf("ens160 invalid output"))
	}
}

func (s *ENS160Source) compensate(now time.Time) error {
	if s.compensation == nil {
		return nil
	}

	readings := s.compensation.Read(now)
	if readings.Temperature.State != types.ReadingReady || readings.Humidity.State != types.ReadingReady {
		return nil
	}

	temperature := int32(math.Round(readings.Temperature.Value * 1000))

	humidity := int32(math.Round(readings.Humidity.Value * 1000))
	if err := s.device.SetEnvDataMilli(temperature, humidity); err != nil {
		return fmt.Errorf("ens160 compensation: %w", err)
	}

	return nil
}

func openENS160(
	cfg config.Sensor,
	bus drivers.I2C,
	compensation CompensationReader,
	schedule Schedule,
) (*ENS160Source, error) {
	device := ens160.New(bus, cfg.I2CAddress())
	if !device.Connected() {
		return nil, fmt.Errorf("ens160 not connected")
	}

	if err := device.Configure(); err != nil {
		return nil, fmt.Errorf("ens160 configure: %w", err)
	}

	return NewENS160Source(device, compensation, schedule), nil
}
