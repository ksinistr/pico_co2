package display

import (
	"fmt"
	"math"
	"pico_co2/internal/types"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"

	"tinygo.org/x/drivers"
)

func DebugTemp(display drivers.Displayer, r *types.Readings) {
	if display == nil {
		return
	}

	widget.Clear(display)

	sf := font.New(display, font.ProggySZ8)

	width, _ := display.Size()

	tempEnsStr := fmt.Sprintf(
		"T ENS: %.0f%%",
		math.Round(float64(r.Raw.Temperature)),
	)
	humStr := ""
	tempSCDStr := fmt.Sprintf(
		"T SCD: %.0f%%",
		math.Round(float64(r.Raw.Temperature)),
	)
	layout.Three(display, sf, 0, tempEnsStr, humStr, tempSCDStr)

	y := int16(8)

	sf.Print(width-sf.Width("c"), y+9, "c")

	if err := display.Display(); err != nil {
		return
	}
}
