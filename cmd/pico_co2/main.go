package main

import (
	"machine"
	"pico_co2/internal/app"
	"pico_co2/internal/display"
	"pico_co2/internal/sensors"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers/ds3231"
	"tinygo.org/x/drivers/ssd1306"
)

const (
	loopInterval = 50 * time.Millisecond
)

func main() {
	if err := machine.I2C0.Configure(machine.I2CConfig{Frequency: 400 * machine.KHz, SDA: machine.GP4, SCL: machine.GP5}); err != nil {
		println("i2c init:", err.Error())
		return
	}

	device := ssd1306.NewI2C(machine.I2C0)
	device.Configure(ssd1306.Config{Width: 128, Height: 32, Address: ssd1306.Address_128_32})
	device.Command(ssd1306.SETCONTRAST)
	device.Command(0x01)
	device.Command(ssd1306.SETPRECHARGE)
	device.Command(0xE1)
	device.Command(ssd1306.SETVCOMDETECT)
	device.Command(0x30)

	clock := ds3231.New(machine.I2C0)
	if ok := clock.Configure(); !ok {
		display.RenderError(&device, "failed to configure DS3231 sensor")
		return
	}
	initializeClock(&clock)

	left := NewTouchButton(machine.GP10)
	right := NewTouchButton(machine.GP11)
	schedule := app.DefaultSchedule()
	aht, err := sensors.NewAHT20(machine.I2C0)
	if err != nil {
		display.RenderError(&device, err.Error())
		return
	}
	scd, err := sensors.NewSCD4x(machine.I2C0)
	if err != nil {
		display.RenderError(&device, err.Error())
		return
	}
	registrations := []app.Registration{
		{
			ID:       "aht20",
			Priority: 2,
			Capabilities: []types.MeasurementType{
				types.MeasurementTypeTemperature, types.MeasurementTypeHumidity,
			},
			Reader:          aht,
			Interval:        schedule.SensorReadInterval,
			StartupInterval: schedule.StartupReadInterval,
			StartupWindow:   schedule.StartupWindow,
		},
		{
			ID:       "scd4x",
			Priority: 0,
			Capabilities: []types.MeasurementType{
				types.MeasurementTypeCO2, types.MeasurementTypeTemperature, types.MeasurementTypeHumidity,
			},
			Reader:          scd,
			Interval:        schedule.SensorReadInterval,
			StartupInterval: schedule.StartupReadInterval,
			StartupWindow:   schedule.StartupWindow,
		},
		{
			ID:           "ds3231",
			Priority:     1,
			Capabilities: []types.MeasurementType{types.MeasurementTypeTime},
			Reader:       sensors.Clock{Source: &clock},
			Interval:     schedule.TimeReadInterval,
		},
	}

	monitor := app.New(app.Deps{
		Display:      &device,
		Screens:      display.ActiveScreens(),
		RTC:          &clock,
		Sensors:      registrations,
		LeftPressed:  left.Consume,
		RightPressed: right.Consume,
		Schedule:     schedule,
	}, types.DefaultHistoryCapacity)

	watchdog := machine.Watchdog
	watchdog.Configure(machine.WatchdogConfig{TimeoutMillis: machine.WatchdogMaxTimeout})
	watchdog.Start()
	for {
		watchdog.Update()
		monitor.Tick(time.Now())
		time.Sleep(loopInterval)
	}
}

func initializeClock(clock *ds3231.Device) {
	current, _ := clock.ReadTime()
	if current.Year() > 2200 || current.Year() < 2024 {
		fallback := time.Date(2025, 12, 21, 14, 35, 0, 0, time.UTC)
		clock.SetTime(fallback)
		println("DS3231 time set to:", fallback.Format(time.DateTime))
	}
	if !clock.IsRunning() {
		if err := clock.SetRunning(true); err != nil {
			println("ds3231 set running:", err.Error())
		}
	}
}
