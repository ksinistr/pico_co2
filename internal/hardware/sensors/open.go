package sensors

import (
	"fmt"
	"pico_co2/internal/app"
	"pico_co2/internal/config"

	"tinygo.org/x/drivers"
)

func New(
	cfg config.Sensor,
	bus drivers.I2C,
	compensation app.Sensors,
	schedule Schedule,
) (app.Sensors, error) {
	switch cfg.Model {
	case config.AHT20:
		return openAHT20(cfg, bus, schedule), nil
	case config.SCD40:
		return openSCD40(cfg, bus, schedule)
	case config.ENS160:
		return openENS160(cfg, bus, compensation, schedule)
	default:
		return nil, fmt.Errorf("unsupported sensor model %d", cfg.Model)
	}
}
