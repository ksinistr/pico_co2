package config

import "fmt"

func (c Config) Validate() error {
	if err := c.validateSources(); err != nil {
		return err
	}

	if err := c.validateModels(); err != nil {
		return err
	}

	if err := c.validatePins(); err != nil {
		return err
	}

	return c.validateConnections()
}

func (c Config) validateSources() error {
	if len(c.Sensors()) == 0 {
		return fmt.Errorf("config: at least one reading source is required")
	}

	if c.Temperature != nil && !c.Temperature.Model.supportsClimate() {
		return fmt.Errorf("config: temperature source does not measure temperature")
	}

	if c.Humidity != nil && !c.Humidity.Model.supportsClimate() {
		return fmt.Errorf("config: humidity source does not measure humidity")
	}

	if c.CO2 != nil && c.CO2.Model != SCD40 && c.CO2.Model != ENS160 {
		return fmt.Errorf("config: CO2 source does not measure CO2")
	}

	return nil
}

func (c Config) validateModels() error {
	if c.Display.Model != SSD1306 || c.Display.Width != 128 || c.Display.Height != 32 {
		return fmt.Errorf("config: supported display is SSD1306 128x32")
	}

	if err := c.Display.validateConnection(); err != nil {
		return err
	}

	if c.RTC != nil && c.RTC.Model != DS3231 {
		return fmt.Errorf("config: unsupported RTC model %d", c.RTC.Model)
	}

	for _, sensor := range c.Sensors() {
		if err := sensor.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (s Sensor) validate() error {
	switch s.Model {
	case AHT20, SCD40, ENS160:
	default:
		return fmt.Errorf("config: unsupported sensor model %d", s.Model)
	}

	if s.Interval < 0 {
		return fmt.Errorf("config: sensor interval must not be negative")
	}

	if s.Compensation == nil {
		return nil
	}

	if s.Model != ENS160 || !s.Compensation.Model.supportsClimate() {
		return fmt.Errorf("config: only ENS160 accepts compensation from a temperature and humidity sensor")
	}

	return nil
}

type busAddress struct {
	bus     I2CBus
	address uint16
}

func (c Config) validateConnections() error {
	if c.Display.SPI != nil && !c.hasSPI(c.Display.SPI.Bus) {
		return fmt.Errorf("config: SPI controller %d is not configured", c.Display.SPI.Bus)
	}

	used := make(map[busAddress]bool)

	if c.Display.I2C != nil {
		connection := c.Display.I2C
		if err := c.claimAddress(used, connection.Bus, connection.I2CAddress()); err != nil {
			return err
		}
	}

	if c.RTC != nil {
		if err := c.claimAddress(used, c.RTC.Bus, c.RTC.I2CAddress()); err != nil {
			return err
		}
	}

	for _, sensor := range c.Sensors() {
		if err := c.claimAddress(used, sensor.Bus, sensor.I2CAddress()); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) claimAddress(used map[busAddress]bool, bus I2CBus, address uint16) error {
	if !c.hasI2C(bus) {
		return fmt.Errorf("config: I2C controller %d is not configured", bus)
	}

	key := busAddress{bus: bus, address: address}
	if address < 0x08 || address > 0x77 || used[key] {
		return fmt.Errorf("config: invalid or conflicting I2C address 0x%02x on bus %d", address, bus)
	}

	used[key] = true

	return nil
}

func (d Display) validateConnection() error {
	if (d.I2C == nil) == (d.SPI == nil) {
		return fmt.Errorf("config: display needs exactly one I2C or SPI connection")
	}

	return nil
}

func (c Config) hasI2C(bus I2CBus) bool {
	for _, candidate := range c.I2C {
		if candidate.Controller == bus {
			return true
		}
	}

	return false
}

func (c Config) hasSPI(bus SPIBus) bool {
	for _, candidate := range c.SPI {
		if candidate.Controller == bus {
			return true
		}
	}

	return false
}
