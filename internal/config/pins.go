package config

import "fmt"

func (c Config) validatePins() error {
	pins := make(map[GPIO]string)
	if err := c.claimI2CPins(pins); err != nil {
		return err
	}

	if err := c.claimSPIPins(pins); err != nil {
		return err
	}

	if c.Display.SPI != nil {
		connection := c.Display.SPI
		if err := claimPins(pins, "SPI display", connection.CS, connection.DC, connection.Reset); err != nil {
			return err
		}
	}

	if c.Buttons != nil {
		return claimPins(pins, "buttons", c.Buttons.Left, c.Buttons.Right)
	}

	return nil
}

func (c Config) claimI2CPins(pins map[GPIO]string) error {
	controllers := make(map[I2CBus]bool)
	for _, bus := range c.I2C {
		if bus.Controller > I2C1 || controllers[bus.Controller] {
			return fmt.Errorf("config: invalid or repeated I2C controller %d", bus.Controller)
		}

		controllers[bus.Controller] = true
		if err := claimPins(pins, "I2C bus", bus.SDA, bus.SCL); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) claimSPIPins(pins map[GPIO]string) error {
	controllers := make(map[SPIBus]bool)
	for _, bus := range c.SPI {
		if bus.Controller > SPI1 || controllers[bus.Controller] {
			return fmt.Errorf("config: invalid or repeated SPI controller %d", bus.Controller)
		}

		if bus.Mode > Mode3 {
			return fmt.Errorf("config: invalid SPI mode %d", bus.Mode)
		}

		controllers[bus.Controller] = true
		if err := claimPins(pins, "SPI bus", bus.SCK, bus.MOSI, bus.MISO); err != nil {
			return err
		}
	}

	return nil
}

func claimPins(pins map[GPIO]string, use string, claimed ...GPIO) error {
	for _, pin := range claimed {
		if previous, found := pins[pin]; found {
			return fmt.Errorf("config: GPIO %d is used by %s and %s", pin, previous, use)
		}

		pins[pin] = use
	}

	return nil
}
