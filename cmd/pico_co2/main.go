package main

import (
	"pico_co2/internal/config"
	"pico_co2/internal/hardware/pico"
)

func main() {
	scd40 := &config.Sensor{Model: config.SCD40, Bus: config.I2C0}
	cfg := config.Config{
		I2C: []config.I2C{
			{Controller: config.I2C0, SDA: 4, SCL: 5, Frequency: 400000},
		},
		Display: config.Display{
			Model:  config.SSD1306,
			I2C:    &config.I2CDisplay{Bus: config.I2C0},
			Width:  128,
			Height: 32,
		},
		Temperature: scd40,
		Humidity:    scd40,
		CO2:         scd40,
		RTC:         &config.RTC{Model: config.DS3231, Bus: config.I2C0},
		Buttons:     &config.Buttons{Left: 10, Right: 11},
	}

	if err := pico.Run(cfg); err != nil {
		println("device init:", err.Error())
	}
}
