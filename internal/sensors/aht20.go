package sensors

import (
	"fmt"
	"pico_co2/internal/types"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/aht20"
)

type SensorAHT20 struct {
	aht20 *aht20.Device
}

func NewAHT20(bus drivers.I2C) (*SensorAHT20, error) {
	aht := aht20.New(bus)
	aht.Reset()
	aht.Configure()

	return &SensorAHT20{
		aht20: &aht,
	}, nil
}

func (d *SensorAHT20) Read() ([]types.Measurement, error) {
	m := make([]types.Measurement, 0, 3)

	if err := d.aht20.Read(); err != nil {
		return nil, fmt.Errorf("aht20 read: %w", err)
	}

	m = append(m, types.Measurement{
		Type:  types.MeasurementTypeTemperature,
		Value: float64(d.aht20.Celsius()),
	})

	m = append(m, types.Measurement{
		Type:  types.MeasurementTypeHumidity,
		Value: float64(d.aht20.RelHumidity()),
	})

	return m, nil
}
