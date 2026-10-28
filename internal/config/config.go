package config

import "time"

type I2CBus uint8

const (
	I2C0 I2CBus = iota
	I2C1
)

type SPIBus uint8

const (
	SPI0 SPIBus = iota
	SPI1
)

type SPIMode uint8

const (
	Mode0 SPIMode = iota
	Mode1
	Mode2
	Mode3
)

type GPIO uint8

type SensorModel uint8

const (
	AHT20 SensorModel = iota + 1
	SCD40
	ENS160
)

type DisplayModel uint8

const SSD1306 DisplayModel = 1

type RTCModel uint8

const DS3231 RTCModel = 1

type I2C struct {
	Controller I2CBus
	SDA        GPIO
	SCL        GPIO
	Frequency  uint32
}

type SPI struct {
	Controller SPIBus
	SCK        GPIO
	MOSI       GPIO
	MISO       GPIO
	Frequency  uint32
	Mode       SPIMode
	LSBFirst   bool
}

type Sensor struct {
	Model        SensorModel
	Bus          I2CBus
	Address      uint16
	Interval     time.Duration
	Compensation *Sensor
}

type I2CDisplay struct {
	Bus     I2CBus
	Address uint16
}

type SPIDisplay struct {
	Bus   SPIBus
	CS    GPIO
	DC    GPIO
	Reset GPIO
}

type Display struct {
	Model  DisplayModel
	Width  int16
	Height int16
	I2C    *I2CDisplay
	SPI    *SPIDisplay
}

type RTC struct {
	Model   RTCModel
	Bus     I2CBus
	Address uint16
}

type Buttons struct {
	Left  GPIO
	Right GPIO
}

type Config struct {
	I2C         []I2C
	SPI         []SPI
	Display     Display
	Temperature *Sensor
	Humidity    *Sensor
	CO2         *Sensor
	RTC         *RTC
	Buttons     *Buttons
}

func (c Config) Sensors() []*Sensor {
	result := make([]*Sensor, 0, 3)
	seen := make(map[*Sensor]bool)

	var visit func(*Sensor)

	visit = func(sensor *Sensor) {
		if sensor == nil || seen[sensor] {
			return
		}

		seen[sensor] = true
		visit(sensor.Compensation)
		result = append(result, sensor)
	}
	visit(c.Temperature)
	visit(c.Humidity)
	visit(c.CO2)

	return result
}

func (s Sensor) I2CAddress() uint16 {
	if s.Address != 0 {
		return s.Address
	}

	switch s.Model {
	case AHT20:
		return 0x38
	case SCD40:
		return 0x62
	case ENS160:
		return 0x53
	default:
		return 0
	}
}

func (d I2CDisplay) I2CAddress() uint16 {
	if d.Address != 0 {
		return d.Address
	}

	return 0x3c
}

func (r RTC) I2CAddress() uint16 {
	if r.Address != 0 {
		return r.Address
	}

	return 0x68
}

func (m SensorModel) supportsClimate() bool {
	return m == AHT20 || m == SCD40
}
