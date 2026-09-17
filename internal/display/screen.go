package display

import (
	"pico_co2/internal/types"
	"tinygo.org/x/drivers"
)

type Screen struct {
	ID     string
	Render func(drivers.Displayer, *types.Readings)
}
