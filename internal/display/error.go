package display

import (
	"pico_co2/internal/types"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"

	"tinygo.org/x/drivers"
)

// RenderError fills the screen with a wrapped error message.
//
//	scd4x read: i2c
//	timeout while
//	reading sensor
func RenderError(display drivers.Displayer, r *types.Readings) {
	if display == nil {
		return
	}

	widget.Clear(display)
	face := font.New(display, font.ProggySZ8)

	if r != nil && r.Error != "" {
		layout.LongText(display, face, 0, 0, r.Error)
	} else {
		layout.LongText(display, face, 0, 0, "No error message available")
	}

	if err := display.Display(); err != nil {
		return
	}
}
