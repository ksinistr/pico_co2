//go:build tinygo

package display

import (
	"fmt"
	"machine"
	"pico_co2/internal/config"
	"pico_co2/internal/device"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"
)

type Driver struct {
	pin func(config.GPIO) (machine.Pin, error)
}

func New(pin func(config.GPIO) (machine.Pin, error)) Driver {
	return Driver{pin: pin}
}

func (d Driver) Open(cfg config.Display, ports device.DisplayPorts) (drivers.Displayer, error) {
	var display *ssd1306.Device
	settings := ssd1306.Config{Width: cfg.Width, Height: cfg.Height}
	if cfg.I2C != nil {
		display = ssd1306.NewI2C(ports.I2C)
		settings.Address = cfg.I2C.I2CAddress()
	} else {
		created, err := d.newSPIDisplay(*cfg.SPI, ports.SPI)
		if err != nil {
			return nil, err
		}
		display = created
	}
	display.Configure(settings)
	display.Command(ssd1306.SETCONTRAST)
	display.Command(0x01)
	display.Command(ssd1306.SETPRECHARGE)
	display.Command(0xE1)
	display.Command(ssd1306.SETVCOMDETECT)
	display.Command(0x30)

	return display, nil
}

func (d Driver) newSPIDisplay(cfg config.SPIDisplay, bus drivers.SPI) (*ssd1306.Device, error) {
	dc, err := d.pin(cfg.DC)
	if err != nil {
		return nil, fmt.Errorf("display DC: %w", err)
	}
	reset, err := d.pin(cfg.Reset)
	if err != nil {
		return nil, fmt.Errorf("display reset: %w", err)
	}
	cs, err := d.pin(cfg.CS)
	if err != nil {
		return nil, fmt.Errorf("display CS: %w", err)
	}

	return ssd1306.NewSPI(bus, dc, reset, cs), nil
}
