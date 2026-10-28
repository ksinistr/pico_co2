package rtc

import (
	"fmt"
	"pico_co2/internal/config"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ds3231"
)

func New(cfg config.RTC, bus drivers.I2C) (RTC, error) {
	clock := ds3231.New(bus)

	clock.Address = cfg.I2CAddress()
	if !clock.Configure() {
		return nil, fmt.Errorf("configure ds3231")
	}

	initializeClock(&clock)

	return &clock, nil
}

func initializeClock(clock *ds3231.Device) {
	current, err := clock.ReadTime()
	if err != nil {
		println("ds3231 read:", err.Error())
		return
	}

	if current.Year() > 2200 || current.Year() < 2024 {
		fallback := time.Date(2025, 12, 21, 14, 35, 0, 0, time.UTC)
		if err := clock.SetTime(fallback); err != nil {
			println("ds3231 set time:", err.Error())
		}
	}

	if !clock.IsRunning() {
		if err := clock.SetRunning(true); err != nil {
			println("ds3231 set running:", err.Error())
		}
	}
}
