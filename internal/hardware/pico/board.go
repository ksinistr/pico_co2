//go:build tinygo

package pico

import (
	"fmt"
	"machine"
	"pico_co2/internal/config"

	"tinygo.org/x/drivers"
)

type board struct{}

func (board) NewI2C(cfg config.I2C) (drivers.I2C, error) {
	sda, err := gpio(cfg.SDA)
	if err != nil {
		return nil, err
	}
	scl, err := gpio(cfg.SCL)
	if err != nil {
		return nil, err
	}
	bus := machine.I2C0
	if cfg.Controller == config.I2C1 {
		bus = machine.I2C1
	}
	frequency := cfg.Frequency
	if frequency == 0 {
		frequency = 400 * machine.KHz
	}
	if err := bus.Configure(machine.I2CConfig{Frequency: frequency, SDA: sda, SCL: scl}); err != nil {
		return nil, fmt.Errorf("i2c configure: %w", err)
	}

	return bus, nil
}

func (board) NewSPI(cfg config.SPI) (drivers.SPI, error) {
	sck, err := gpio(cfg.SCK)
	if err != nil {
		return nil, err
	}
	mosi, err := gpio(cfg.MOSI)
	if err != nil {
		return nil, err
	}
	miso, err := gpio(cfg.MISO)
	if err != nil {
		return nil, err
	}
	bus := machine.SPI0
	if cfg.Controller == config.SPI1 {
		bus = machine.SPI1
	}
	if err := bus.Configure(machine.SPIConfig{
		Frequency: cfg.Frequency,
		Mode:      uint8(cfg.Mode),
		LSBFirst:  cfg.LSBFirst,
		SCK:       sck, SDO: mosi, SDI: miso,
	}); err != nil {
		return nil, fmt.Errorf("spi configure: %w", err)
	}

	return bus, nil
}

func (board) NewButton(pin config.GPIO) (func() bool, error) {
	configured, err := gpio(pin)
	if err != nil {
		return nil, err
	}

	return NewTouchButton(configured).Consume, nil
}

func gpio(pin config.GPIO) (machine.Pin, error) {
	if pin > 29 {
		return 0, fmt.Errorf("Pico GPIO %d is outside 0..29", pin)
	}

	return machine.Pin(pin), nil
}
