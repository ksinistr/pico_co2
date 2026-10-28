package sensors

import (
	"fmt"
	"pico_co2/internal/config"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/aht20"
)

type AHT20Device interface {
	Read() error
	Celsius() float32
	RelHumidity() float32
}

type AHT20Source struct {
	device   AHT20Device
	readings types.Readings
	timer    readTimer
}

func NewAHT20Source(device AHT20Device, schedule Schedule) *AHT20Source {
	return &AHT20Source{device: device, timer: newReadTimer(schedule)}
}

func (s *AHT20Source) Read(now time.Time) types.Readings {
	if !s.timer.due(now) {
		return s.readings
	}

	if err := s.device.Read(); err != nil {
		err = fmt.Errorf("aht20 read: %w", err)
		s.readings.Temperature = failed(s.readings.Temperature, err)
		s.readings.Humidity = failed(s.readings.Humidity, err)

		return s.readings
	}

	s.readings.Temperature = ready(s.readings.Temperature, float64(s.device.Celsius()), now)
	s.readings.Humidity = ready(s.readings.Humidity, float64(s.device.RelHumidity()), now)

	return s.readings
}

func openAHT20(cfg config.Sensor, bus drivers.I2C, schedule Schedule) *AHT20Source {
	device := aht20.New(bus)
	device.Address = cfg.I2CAddress()
	device.Reset()
	device.Configure()

	return NewAHT20Source(&device, schedule)
}
