package display

import (
	"pico_co2/internal/types"
	"tinygo.org/x/drivers"
)

// Screen describes one selectable rendering mode.
type Screen struct {
	ID string
	// Render cannot report display errors; the hardware loop relies on the watchdog for recovery.
	Render          func(drivers.Displayer, *types.Readings)
	AllowsClockEdit bool
}
