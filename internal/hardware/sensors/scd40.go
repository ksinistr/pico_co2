package sensors

import (
	"fmt"
	"pico_co2/internal/config"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/scd4x"
)

type SCD4xDevice interface {
	DataReady() (bool, error)
	ReadCO2() (int32, error)
	ReadTemperature() (int32, error)
	ReadHumidity() (int32, error)
}

type SCD4xSource struct {
	device   SCD4xDevice
	readings types.Readings
	timer    readTimer
}

func NewSCD4xSource(device SCD4xDevice, schedule Schedule) *SCD4xSource {
	return &SCD4xSource{device: device, timer: newReadTimer(schedule)}
}

func (s *SCD4xSource) Read(now time.Time) types.Readings {
	if !s.timer.due(now) {
		return s.readings
	}

	available, err := s.device.DataReady()
	if err != nil {
		s.fail(fmt.Errorf("scd4x ready: %w", err))
		return s.readings
	}

	if !available {
		if s.readings.CO2.State == types.ReadingUnavailable {
			s.readings = types.Readings{
				Temperature: warming(s.readings.Temperature),
				Humidity:    warming(s.readings.Humidity),
				CO2:         warming(s.readings.CO2),
			}
		}

		return s.readings
	}

	s.readSample(now)

	return s.readings
}

func (s *SCD4xSource) readSample(now time.Time) {
	co2, err := s.device.ReadCO2()
	if err != nil {
		s.fail(fmt.Errorf("scd4x co2: %w", err))
		return
	}

	temperature, err := s.device.ReadTemperature()
	if err != nil {
		s.fail(fmt.Errorf("scd4x temperature: %w", err))
		return
	}

	humidity, err := s.device.ReadHumidity()
	if err != nil {
		s.fail(fmt.Errorf("scd4x humidity: %w", err))
		return
	}

	if co2 <= 0 {
		s.readings.Temperature = warming(s.readings.Temperature)
		s.readings.Humidity = warming(s.readings.Humidity)
		s.readings.CO2 = warming(s.readings.CO2)

		return
	}

	s.readings = types.Readings{
		Temperature: ready(s.readings.Temperature, float64(temperature)/1000, now),
		Humidity:    ready(s.readings.Humidity, float64(humidity), now),
		CO2:         ready(s.readings.CO2, float64(co2), now),
	}
}

func (s *SCD4xSource) fail(err error) {
	s.readings.Temperature = failed(s.readings.Temperature, err)
	s.readings.Humidity = failed(s.readings.Humidity, err)
	s.readings.CO2 = failed(s.readings.CO2, err)
}

func openSCD40(cfg config.Sensor, bus drivers.I2C, schedule Schedule) (*SCD4xSource, error) {
	device := scd4x.New(bus)

	device.Address = uint8(cfg.I2CAddress())
	if err := initializeSCD4x(device); err != nil {
		return nil, err
	}

	return NewSCD4xSource(device, schedule), nil
}

func initializeSCD4x(device *scd4x.Device) error {
	time.Sleep(1500 * time.Millisecond)

	if err := device.Configure(); err != nil {
		return fmt.Errorf("scd4x configure: %w", err)
	}

	time.Sleep(1500 * time.Millisecond)

	if err := device.StartPeriodicMeasurement(); err != nil {
		return fmt.Errorf("scd4x start: %w", err)
	}

	time.Sleep(1500 * time.Millisecond)

	return nil
}
