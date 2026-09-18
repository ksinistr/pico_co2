// Package widget draws reusable figures on a monochrome display.
// Coordinates start at the top-left corner and the y-axis grows downward.
//
//	(0,0)                         (width,0)
//	  +-------------------------------+
//	  |                               |
//	  +-------------------------------+
//	(0,height)               (width,height)
package widget

import (
	"fmt"
	"image/color"
	"pico_co2/pkg/font"
	"pico_co2/pkg/sparkline"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinydraw"
)

// TrendDirection selects the symbol drawn by Trend.
type TrendDirection uint8

const (
	TrendStable TrendDirection = iota
	TrendRising
	TrendFalling
	TrendUnknown
)

// Clear fills the complete display with black.
//
//	+-------------------------------+
//	|                               |
//	+-------------------------------+
func Clear(display drivers.Displayer) {
	if display == nil {
		return
	}

	width, height := display.Size()
	if err := tinydraw.FilledRectangle(display, 0, 0, width, height, blackColor()); err != nil {
		return
	}
}

// TwoSideBar draws a label with dot scales on either side. A positive value
// fills dots on the right; a negative value fills dots on the left. Large
// circles are filled and small circles are empty positions.
//
//	left=0, right=2, value=1    left=3, right=3, value=-2
//
//	H  O  .  O  T  .
//	^  ^                        ^        ^
//	x  first right dot          x        label
//
// It returns the x-coordinate immediately after the right scale.
func TwoSideBar(display drivers.Displayer, face font.Face, x, y, value int16, label string, left, right int16) int16 {
	if display == nil {
		return x
	}

	const radius, spacing int16 = 3, 5

	barY := y + radius + 1
	for i := range left {
		barX := x + radius + i*(2*radius+spacing)

		r := int16(1)
		if value < 0 && i >= left+value {
			r = radius
		}

		tinydraw.FilledCircle(display, barX, barY, r, whiteColor())
	}

	if left > 0 {
		x += left*(2*radius+spacing) + radius
	}

	labelWidth := face.Print(x, y+radius+1-face.Height()/2, label)

	x += labelWidth + radius + spacing
	for i := range right {
		barX := x + radius + i*(2*radius+spacing)

		r := int16(1)
		if value > 0 && i < value {
			r = radius
		}

		tinydraw.FilledCircle(display, barX, barY, r, whiteColor())
	}

	if right == 0 {
		return x
	}

	return x + right*(2*radius+spacing) + radius
}

// Trend draws an 8-pixel-wide direction symbol centered on (x,y).
// TrendUnknown draws nothing.
//
//	rising     falling     stable
//	   /\          --         ----
//	  /__\         \/
func Trend(display drivers.Displayer, x, y int16, direction TrendDirection) {
	const size int16 = 4

	switch direction {
	case TrendRising:
		tinydraw.FilledTriangle(display, x-size, y+size, x+size, y+size, x, y-size, whiteColor())
	case TrendFalling:
		tinydraw.FilledTriangle(display, x-size, y-size, x+size, y-size, x, y+size, whiteColor())
	case TrendStable:
		if err := tinydraw.FilledRectangle(display, x-size, y-1, 2*size, 2, whiteColor()); err != nil {
			return
		}
	case TrendUnknown:
	}
}

// ScaleX maps a value onto a horizontal scale and clamps it to both ends.
//
//	min                 value                 max
//	 |--------------------|--------------------|
//	 0                   result              width
func ScaleX(width int16, value, lower, upper float32) int16 {
	if width <= 0 || upper <= lower {
		return 0
	}

	if value < lower {
		value = lower
	}

	if value > upper {
		value = upper
	}

	return int16(float32(width) * (value - lower) / (upper - lower))
}

// ThermalRange defines the boundaries shown by ThermalScale.
type ThermalRange struct {
	Min  float32
	Warm float32
	Hot  float32
	Max  float32
}

// ThermalScale draws a full-width heat scale. Below Warm only the baseline is
// visible; above it the bar grows toward Value. Hot is marked by a black gap.
// Values outside Min..Max are clamped to the display edges.
//
//	Min          Warm                    Hot         Max
//	 |____________|#######################|#|_________|
//	              |<------ Value -------->|
//	 0                                           width
func ThermalScale(display drivers.Displayer, y, height int16, value float32, scale ThermalRange) {
	if display == nil {
		return
	}

	width, _ := display.Size()
	origin := ScaleX(width, scale.Warm, scale.Min, scale.Max)
	// Draw the thin baseline from the left edge to the Warm threshold.
	if err := tinydraw.FilledRectangle(display, 0, y+height-1, origin, 1, whiteColor()); err != nil {
		return
	}
	// Mark the Warm threshold with a vertical divider extending above the bar.
	if err := tinydraw.FilledRectangle(display, origin-1, y, 1, height, blackColor()); err != nil {
		return
	}

	valueX := ScaleX(width, value, scale.Min, scale.Max)
	if valueX <= origin {
		return
	}
	// Fill the bar between the Warm threshold and the current value.
	if err := tinydraw.FilledRectangle(display, origin, y, valueX-origin+1, height, whiteColor()); err != nil {
		return
	}

	hotX := ScaleX(width, scale.Hot, scale.Min, scale.Max)
	// Mark the Hot threshold with a one-pixel white divider and black gaps.
	if err := tinydraw.FilledRectangle(display, hotX-1, y, 1, height, blackColor()); err != nil {
		return
	}

	if err := tinydraw.FilledRectangle(display, hotX, y, 1, height, whiteColor()); err != nil {
		return
	}
}

// Gauge draws a proportional horizontal fill over a baseline track. Values
// outside min..max are clamped to the ends.
//
//	min          value                    max
//	 |###############|______________________|
//	 x                                 x+width
func Gauge(display drivers.Displayer, x, y, width, height int16, value, lower, upper float32) {
	if display == nil || width <= 0 || height <= 0 {
		return
	}

	if err := tinydraw.FilledRectangle(display, x, y+height-1, width, 1, whiteColor()); err != nil {
		return
	}

	filled := ScaleX(width, value, lower, upper)
	if filled <= 0 {
		return
	}

	if err := tinydraw.FilledRectangle(display, x, y, filled, height, whiteColor()); err != nil {
		return
	}
}

func GaugeBordered(display drivers.Displayer, x, y, width, height int16, value, lower, upper float32) {
	if display == nil || width <= 0 || height <= 0 {
		return
	}

	if err := tinydraw.Rectangle(display, x, y, width, height, whiteColor()); err != nil {
		return
	}

	filled := ScaleX(width, value, lower, upper)
	if filled <= 0 {
		return
	}

	if err := tinydraw.FilledRectangle(display, x, y, filled, height, whiteColor()); err != nil {
		return
	}
}

// GaugeMark draws a threshold divider inside a Gauge: a black gap keeps it
// visible once the fill passes it.
//
//	|#########|#|__________|
//	          ^
//	        value
func GaugeMark(display drivers.Displayer, x, y, width, height int16, value, lower, upper float32) {
	if display == nil || width <= 0 || height <= 0 {
		return
	}

	markX := x + ScaleX(width, value, lower, upper)
	if err := tinydraw.FilledRectangle(display, markX-1, y, 1, height, blackColor()); err != nil {
		return
	}

	if err := tinydraw.FilledRectangle(display, markX, y, 1, height, whiteColor()); err != nil {
		return
	}
}

// Sparkline draws one bottom-aligned vertical bar per sample. If data is wider
// than the available area, only the newest values are shown.
//
//	y       #       #
//	     #  #  #    #
//	  #  #  #  # #  #
//	y+h------------------ x..x+width
func Sparkline(display drivers.Displayer, x, y, width, height int16, data []int16) {
	if display == nil || width <= 0 || height <= 0 || len(data) == 0 {
		return
	}

	if len(data) > int(width) {
		data = data[len(data)-int(width):]
	}

	values := sparkline.NewSparkline(int(height)).Process(data)
	for i, value := range values {
		if value <= 0 {
			continue
		}

		if err := tinydraw.FilledRectangle(display, x+int16(i), y+height-value, 1, value, whiteColor()); err != nil {
			return
		}
	}
}

// SquareBar draws four horizontal blocks and fills the first value blocks.
// A zero value draws nothing.
// It is retained as a reusable gallery figure; active screens do not use it.
//
//	value=3
//	+-----+  +-----+  +-----+  +-----+
//	|#####|  |#####|  |#####|  |     |
//	+-----+  +-----+  +-----+  +-----+
func SquareBar(display drivers.Displayer, x, y int16, value uint8) {
	if value == 0 {
		return
	}

	const width, height, gap int16 = (128 - 6*3) / 4, 9, 6
	for i := range int16(4) {
		if err := drawSquare(display, x, y, width, height, i < int16(value)); err != nil {
			return
		}

		x += width + gap
	}
}

func drawSquare(display drivers.Displayer, x, y, width, height int16, filled bool) error {
	if filled {
		if err := tinydraw.FilledRectangle(display, x, y, width, height, whiteColor()); err != nil {
			return fmt.Errorf("fill square: %w", err)
		}

		return nil
	}

	if err := tinydraw.Rectangle(display, x, y, width, height, whiteColor()); err != nil {
		return fmt.Errorf("draw square: %w", err)
	}

	return nil
}

// VerticalBar draws a bottom-filled stack of two-pixel blocks.
// It is retained as a reusable gallery figure; active screens do not use it.
//
//	height=4, value=3
//	..
//	##
//	##
//	##
func VerticalBar(display drivers.Displayer, x, y, value, height int16) {
	if value < 0 {
		value = 0
	}

	if value > height {
		value = height
	}

	for i := range height {
		on := i >= height-value
		if err := tinydraw.FilledRectangle(display, x, y+i*3, 2, 2, colorFor(on)); err != nil {
			return
		}
	}
}

func colorFor(on bool) color.RGBA {
	if on {
		return whiteColor()
	}

	return blackColor()
}

func whiteColor() color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

func blackColor() color.RGBA {
	return color.RGBA{0, 0, 0, 255}
}
