//go:build tinygo

package pico

import (
	"machine"
	"pico_co2/internal/app"
	"pico_co2/internal/config"
	"pico_co2/internal/device"
	"pico_co2/internal/display"
	hardwaredisplay "pico_co2/internal/hardware/display"
	hardwarertc "pico_co2/internal/hardware/rtc"
	"pico_co2/internal/hardware/sensors"
	"pico_co2/internal/types"
	"time"
)

const loopInterval = 50 * time.Millisecond

func Run(cfg config.Config) error {
	constructors := device.Peripherals{
		NewDisplay: hardwaredisplay.New(gpio).Open,
		NewSensor:  sensors.New,
		NewRTC:     hardwarertc.New,
	}
	deps, err := device.Build(cfg, board{}, constructors)
	if err != nil {
		if deps.Display != nil {
			display.RenderError(deps.Display, err.Error())
		}
		return err
	}

	application := app.New(deps, types.DefaultHistoryCapacity)
	watchdog := machine.Watchdog
	watchdog.Configure(
		machine.WatchdogConfig{TimeoutMillis: machine.WatchdogMaxTimeout},
	)
	watchdog.Start()
	for {
		watchdog.Update()
		application.Tick(time.Now())
		time.Sleep(loopInterval)
	}
}
