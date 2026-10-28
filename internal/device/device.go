package device

import (
	"fmt"
	"pico_co2/internal/app"
	"pico_co2/internal/config"
	"pico_co2/internal/display"
	"pico_co2/internal/hardware/rtc"
	"pico_co2/internal/hardware/sensors"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers"
)

type Platform interface {
	NewI2C(config.I2C) (drivers.I2C, error)
	NewSPI(config.SPI) (drivers.SPI, error)
	NewButton(config.GPIO) (func() bool, error)
}

type DisplayPorts struct {
	I2C drivers.I2C
	SPI drivers.SPI
}

type Peripherals struct {
	NewDisplay func(config.Display, DisplayPorts) (drivers.Displayer, error)
	NewSensor  func(config.Sensor, drivers.I2C, app.Sensors, sensors.Schedule) (app.Sensors, error)
	NewRTC     func(config.RTC, drivers.I2C) (rtc.RTC, error)
}

func Build(
	cfg config.Config,
	platform Platform,
	peripherals Peripherals,
) (app.Deps, error) {
	deps := app.Deps{Schedule: app.DefaultSchedule()}
	if err := cfg.Validate(); err != nil {
		return deps, fmt.Errorf("device config: %w", err)
	}

	if peripherals.NewDisplay == nil || peripherals.NewSensor == nil {
		return deps, fmt.Errorf(
			"device: display and sensor constructors are required",
		)
	}

	if cfg.RTC != nil && peripherals.NewRTC == nil {
		return deps, fmt.Errorf("device: RTC constructor is required")
	}

	buses, err := openBuses(cfg, platform)
	if err != nil {
		return deps, err
	}

	deps.Display, err = peripherals.NewDisplay(
		cfg.Display,
		buses.display(cfg.Display),
	)
	if err != nil {
		return deps, fmt.Errorf("device display: %w", err)
	}

	deps.Sensors, err = buildSensors(cfg, buses.i2c, peripherals, deps.Schedule)
	if err != nil {
		return deps, err
	}

	if err := openControls(cfg, buses.i2c, platform, peripherals, &deps); err != nil {
		return deps, err
	}

	deps.Screens = display.ActiveScreens()

	return deps, nil
}

type buses struct {
	i2c map[config.I2CBus]drivers.I2C
	spi map[config.SPIBus]drivers.SPI
}

func openBuses(cfg config.Config, platform Platform) (buses, error) {
	result := buses{
		i2c: make(map[config.I2CBus]drivers.I2C),
		spi: make(map[config.SPIBus]drivers.SPI),
	}

	for _, bus := range cfg.I2C {
		opened, err := platform.NewI2C(bus)
		if err != nil {
			return result, fmt.Errorf("device I2C%d: %w", bus.Controller, err)
		}

		result.i2c[bus.Controller] = opened
	}

	for _, bus := range cfg.SPI {
		opened, err := platform.NewSPI(bus)
		if err != nil {
			return result, fmt.Errorf("device SPI%d: %w", bus.Controller, err)
		}

		result.spi[bus.Controller] = opened
	}

	return result, nil
}

func (b buses) display(cfg config.Display) DisplayPorts {
	if cfg.I2C != nil {
		return DisplayPorts{I2C: b.i2c[cfg.I2C.Bus]}
	}

	return DisplayPorts{SPI: b.spi[cfg.SPI.Bus]}
}

func buildSensors(
	cfg config.Config,
	buses map[config.I2CBus]drivers.I2C,
	peripherals Peripherals,
	schedule app.Schedule,
) (*sensorGroup, error) {
	sensorReaders := make(map[*config.Sensor]app.Sensors)

	for _, sensor := range cfg.Sensors() {
		timing := sensors.Schedule{
			Interval:        schedule.SensorReadInterval,
			StartupInterval: schedule.StartupReadInterval,
			StartupWindow:   schedule.StartupWindow,
		}
		if sensor.Interval > 0 {
			timing.Interval = sensor.Interval
		}

		reader, err := peripherals.NewSensor(
			*sensor,
			buses[sensor.Bus],
			sensorReaders[sensor.Compensation],
			timing,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"device sensor model %d: %w",
				sensor.Model,
				err,
			)
		}

		sensorReaders[sensor] = reader
	}

	return &sensorGroup{
		temperature: sensorReaders[cfg.Temperature],
		humidity:    sensorReaders[cfg.Humidity],
		co2:         sensorReaders[cfg.CO2],
	}, nil
}

func openControls(
	cfg config.Config,
	buses map[config.I2CBus]drivers.I2C,
	platform Platform,
	peripherals Peripherals,
	deps *app.Deps,
) error {
	if cfg.RTC != nil {
		clock, err := peripherals.NewRTC(*cfg.RTC, buses[cfg.RTC.Bus])
		if err != nil {
			return fmt.Errorf("device RTC: %w", err)
		}

		deps.Clock = clock
		deps.SetTime = clock.SetTime
	}

	if cfg.Buttons == nil {
		return nil
	}

	left, err := platform.NewButton(cfg.Buttons.Left)
	if err != nil {
		return fmt.Errorf("device left button: %w", err)
	}

	right, err := platform.NewButton(cfg.Buttons.Right)
	if err != nil {
		return fmt.Errorf("device right button: %w", err)
	}

	deps.LeftPressed, deps.RightPressed = left, right

	return nil
}

type sensorGroup struct {
	temperature app.Sensors
	humidity    app.Sensors
	co2         app.Sensors
}

func (s *sensorGroup) Read(now time.Time) types.Readings {
	var result types.Readings
	if s.temperature != nil {
		result.Temperature = s.temperature.Read(now).Temperature
	}

	if s.humidity != nil {
		result.Humidity = s.humidity.Read(now).Humidity
	}

	if s.co2 != nil {
		result.CO2 = s.co2.Read(now).CO2
	}

	return result
}
