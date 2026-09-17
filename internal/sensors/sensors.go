package sensors

import (
	"fmt"
	"time"

	"pico_co2/pkg/ens160"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/aht20"
	"tinygo.org/x/drivers/scd4x"
)

type Raw struct {
	CO2         uint16
	Temperature float32
	Humidity    float32
}

type Reader func() (Raw, error)

type Devices struct {
	aht20  *aht20.Device
	ens160 *ens160.Device
	scd4x  *scd4x.Device
}

func New(bus drivers.I2C) (*Devices, error) {
	aht := aht20.New(bus)
	aht.Reset()
	aht.Configure()

	ens := ens160.New(bus, ens160.DefaultAddress)
	if err := ens.Sleep(); err != nil {
		return nil, fmt.Errorf("ens160 init: %w", err)
	}

	scd := scd4x.New(bus)
	time.Sleep(1500 * time.Millisecond)
	if err := scd.Configure(); err != nil {
		return nil, fmt.Errorf("scd4x configure: %w", err)
	}
	time.Sleep(1500 * time.Millisecond)
	if err := scd.StartPeriodicMeasurement(); err != nil {
		return nil, fmt.Errorf("scd4x start periodic measurement: %w", err)
	}
	time.Sleep(1500 * time.Millisecond)

	return &Devices{aht20: &aht, ens160: ens, scd4x: scd}, nil
}

func (d *Devices) Read() (Raw, error) {
	if err := d.aht20.Read(); err != nil {
		return Raw{}, fmt.Errorf("aht20 read: %w", err)
	}
	co2, err := d.scd4x.ReadCO2()
	if err != nil {
		return Raw{}, fmt.Errorf("scd4x read: %w", err)
	}
	return Raw{CO2: uint16(co2), Temperature: d.aht20.Celsius(), Humidity: d.aht20.RelHumidity()}, nil
}
