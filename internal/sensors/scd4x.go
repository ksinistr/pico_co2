package sensors

import (
	"fmt"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/scd4x"
)

type SensorSCD4x struct {
	scd4x *scd4x.Device
}

func NewSCD4x(bus drivers.I2C) (*SensorSCD4x, error) {
	scd := scd4x.New(bus)
	// Preserve the deployed startup settling windows around SCD4x commands;
	// the driver separately enforces each command's protocol minimum.
	time.Sleep(1500 * time.Millisecond)

	if err := scd.Configure(); err != nil {
		return nil, fmt.Errorf("scd4x configure: %w", err)
	}

	time.Sleep(1500 * time.Millisecond)

	if err := scd.StartPeriodicMeasurement(); err != nil {
		return nil, fmt.Errorf("scd4x start periodic measurement: %w", err)
	}

	time.Sleep(1500 * time.Millisecond)

	return &SensorSCD4x{
		scd4x: scd,
	}, nil
}

func (d *SensorSCD4x) Read() ([]types.Measurement, error) {
	m := make([]types.Measurement, 0, 3)

	co2, err := d.scd4x.ReadCO2()
	if err != nil {
		return nil, fmt.Errorf("scd4x read: %w", err)
	}

	m = append(m, types.Measurement{
		Type:  types.MeasurementTypeCO2,
		Value: float64(co2),
	})

	temp, err := d.scd4x.ReadTemperature()
	if err != nil {
		return nil, fmt.Errorf("scd4x read: %w", err)
	}

	m = append(m, types.Measurement{
		Type:  types.MeasurementTypeTemperature,
		Value: float64(temp) / 1000,
	})

	humid, err := d.scd4x.ReadHumidity()
	if err != nil {
		return nil, fmt.Errorf("scd4x read: %w", err)
	}

	m = append(m, types.Measurement{
		Type:  types.MeasurementTypeHumidity,
		Value: float64(humid),
	})

	return m, nil
}
