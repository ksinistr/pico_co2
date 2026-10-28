package config

import (
	"strings"
	"testing"
	"time"
)

func thermometer() Config {
	sensor := &Sensor{Model: AHT20, Bus: I2C0}
	return Config{
		I2C:         []I2C{{Controller: I2C0, SDA: 4, SCL: 5, Frequency: 400000}},
		Display:     Display{Model: SSD1306, I2C: &I2CDisplay{Bus: I2C0}, Width: 128, Height: 32},
		Temperature: sensor,
		Humidity:    sensor,
	}
}

func TestValidate(t *testing.T) {
	for _, tt := range []struct {
		name    string
		change  func(*Config)
		message string
	}{
		{"shared sensor", func(*Config) {}, ""},
		{"temperature only", func(c *Config) { c.Humidity = nil }, ""},
		{"no sources", func(c *Config) { c.Temperature, c.Humidity = nil, nil }, "reading source"},
		{"CO2 from AHT20", func(c *Config) { c.CO2 = c.Temperature }, "CO2 source"},
		{"temperature from ENS160", func(c *Config) { c.Temperature = &Sensor{Model: ENS160} }, "temperature source"},
		{"humidity from ENS160", func(c *Config) { c.Humidity = &Sensor{Model: ENS160} }, "humidity source"},
		{"unknown sensor", func(c *Config) { c.Temperature.Model = SensorModel(99) }, "temperature source"},
		{"unknown bus", func(c *Config) { c.Temperature.Bus = I2C1 }, "not configured"},
		{"unsupported controller", func(c *Config) { c.I2C[0].Controller = I2CBus(99) }, "controller"},
		{"duplicate controller", func(c *Config) { c.I2C = append(c.I2C, c.I2C[0]) }, "controller"},
		{"same bus pins", func(c *Config) { c.I2C[0].SCL = 4 }, "GPIO"},
		{"button overlaps bus", func(c *Config) { c.Buttons = &Buttons{Left: 4, Right: 11} }, "GPIO"},
		{"buttons overlap", func(c *Config) { c.Buttons = &Buttons{Left: 10, Right: 10} }, "GPIO"},
		{"distinct objects same address", func(c *Config) { c.Humidity = &Sensor{Model: AHT20} }, "conflicting I2C address"},
		{"address on another bus", func(c *Config) {
			c.I2C = append(c.I2C, I2C{Controller: I2C1, SDA: 6, SCL: 7})
			c.Humidity = &Sensor{Model: AHT20, Bus: I2C1}
		}, ""},
		{"invalid address", func(c *Config) { c.Temperature.Address = 0x80 }, "address"},
		{"display conflict", func(c *Config) { c.Display.I2C.Address = 0x38 }, "address"},
		{"unknown display", func(c *Config) { c.Display.Model = DisplayModel(99) }, "display"},
		{"unsupported dimensions", func(c *Config) { c.Display.Height = 64 }, "display"},
		{"unknown RTC", func(c *Config) { c.RTC = &RTC{Model: RTCModel(99)} }, "RTC"},
		{"negative interval", func(c *Config) { c.Temperature.Interval = -time.Second }, "interval"},
		{"unexpected compensation", func(c *Config) { c.Temperature.Compensation = c.Temperature }, "compensation"},
		{"ENS160 compensation", func(c *Config) { c.CO2 = &Sensor{Model: ENS160, Compensation: c.Temperature} }, ""},
		{"cyclic compensation", func(c *Config) { c.CO2 = &Sensor{Model: ENS160}; c.CO2.Compensation = c.CO2 }, "compensation"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := thermometer()
			tt.change(&cfg)
			err := cfg.Validate()
			if tt.message == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("error = %v, want %q", err, tt.message)
			}
		})
	}
}

func TestSensorDependencies(t *testing.T) {
	for _, tt := range []struct {
		name             string
		compensationOnly bool
	}{
		{"shared measurement and compensation source", false},
		{"compensation source without displayed readings", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := thermometer()
			aht20 := cfg.Temperature
			ens160 := &Sensor{Model: ENS160, Compensation: aht20}
			cfg.CO2 = ens160
			if tt.compensationOnly {
				cfg.Temperature, cfg.Humidity = nil, nil
			}
			got := cfg.Sensors()
			if len(got) != 2 || got[0] != aht20 || got[1] != ens160 {
				t.Fatalf("sensor construction order = %v", got)
			}
		})
	}
}

func spiDisplayConfig() Config {
	cfg := thermometer()
	cfg.SPI = []SPI{{Controller: SPI0, SCK: 18, MOSI: 19, MISO: 16, Frequency: 4000000, Mode: Mode0}}
	cfg.Display.I2C = nil
	cfg.Display.SPI = &SPIDisplay{Bus: SPI0, CS: 17, DC: 20, Reset: 21}

	return cfg
}

func TestDisplayConnections(t *testing.T) {
	for _, tt := range []struct {
		name    string
		change  func(*Config)
		message string
	}{
		{"SPI with I2C sensors", func(*Config) {}, ""},
		{"no connection", func(c *Config) { c.Display.SPI = nil }, "exactly one"},
		{"two connections", func(c *Config) { c.Display.I2C = &I2CDisplay{Bus: I2C0} }, "exactly one"},
		{"missing SPI controller", func(c *Config) { c.SPI = nil }, "not configured"},
		{"wrong SPI controller", func(c *Config) { c.Display.SPI.Bus = SPI1 }, "not configured"},
		{"invalid SPI controller", func(c *Config) { c.SPI = append(c.SPI, SPI{Controller: SPIBus(99)}) }, "controller"},
		{"duplicate SPI controller", func(c *Config) { c.SPI = append(c.SPI, c.SPI[0]) }, "controller"},
		{"invalid SPI mode", func(c *Config) { c.SPI[0].Mode = SPIMode(4) }, "mode"},
		{"duplicate SPI bus pins", func(c *Config) { c.SPI[0].MOSI = c.SPI[0].SCK }, "GPIO"},
		{"SPI overlaps I2C", func(c *Config) { c.SPI[0].MISO = c.I2C[0].SDA }, "GPIO"},
		{"display CS overlaps MISO", func(c *Config) { c.Display.SPI.CS = c.SPI[0].MISO }, "GPIO"},
		{"display pins overlap", func(c *Config) { c.Display.SPI.Reset = c.Display.SPI.DC }, "GPIO"},
		{"button overlaps display", func(c *Config) { c.Buttons = &Buttons{Left: 17, Right: 11} }, "GPIO"},
		{"SPI display has no I2C address", func(c *Config) { c.Temperature.Address = 0x3c }, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := spiDisplayConfig()
			tt.change(&cfg)
			err := cfg.Validate()
			if tt.message == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("error = %v, want %q", err, tt.message)
			}
		})
	}
}
