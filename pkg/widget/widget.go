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
	"image/color"

	"pico_co2/pkg/font"
	"pico_co2/pkg/sparkline"
	"tinygo.org/x/drivers"
	"tinygo.org/x/tinydraw"
)

var (
	white = color.RGBA{255, 255, 255, 255}
	black = color.RGBA{0, 0, 0, 255}
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
	tinydraw.FilledRectangle(display, 0, 0, width, height, black)
}

// TwoSideBar draws a label with dot scales on either side. A positive value
// fills dots on the right; a negative value fills dots on the left. Large
// circles are filled and small circles are empty positions.
//
//	left=0, right=2, value=1    left=3, right=3, value=-2
//
//	H  O  .                     .  O  O  T  .  .  .
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
	for i := int16(0); i < left; i++ {
		barX := x + radius + i*(2*radius+spacing)
		r := int16(1)
		if value < 0 && i >= left+value {
			r = radius
		}
		tinydraw.FilledCircle(display, barX, barY, r, white)
	}
	if left > 0 {
		x += left*(2*radius+spacing) + radius
	}
	labelWidth := face.Print(x, y+radius+1-face.Height()/2, label)
	x += labelWidth + radius + spacing
	for i := int16(0); i < right; i++ {
		barX := x + radius + i*(2*radius+spacing)
		r := int16(1)
		if value > 0 && i < value {
			r = radius
		}
		tinydraw.FilledCircle(display, barX, barY, r, white)
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
		tinydraw.FilledTriangle(display, x-size, y+size, x+size, y+size, x, y-size, white)
	case TrendFalling:
		tinydraw.FilledTriangle(display, x-size, y-size, x+size, y-size, x, y+size, white)
	case TrendStable:
		tinydraw.FilledRectangle(display, x-size, y-1, 2*size, 2, white)
	}
}

// ScaleX maps a value onto a horizontal scale and clamps it to both ends.
//
//	min                 value                 max
//	 |--------------------|--------------------|
//	 0                   result              width
func ScaleX(width int16, value, min, max float32) int16 {
	if width <= 0 || max <= min {
		return 0
	}
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}
	return int16(float32(width) * (value - min) / (max - min))
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
	tinydraw.FilledRectangle(display, 0, y+height-1, origin, 1, white)
	// Mark the Warm threshold with a vertical divider extending above the bar.
	tinydraw.FilledRectangle(display, origin-1, y, 1, height, black)
	valueX := ScaleX(width, value, scale.Min, scale.Max)
	if valueX <= origin {
		return
	}
	// Fill the bar between the Warm threshold and the current value.
	tinydraw.FilledRectangle(display, origin, y, valueX-origin+1, height, white)
	hotX := ScaleX(width, scale.Hot, scale.Min, scale.Max)
	// Mark the Hot threshold with a one-pixel white divider and black gaps.
	tinydraw.FilledRectangle(display, hotX-1, y, 1, height, black)
	tinydraw.FilledRectangle(display, hotX, y, 1, height, white)
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
		tinydraw.FilledRectangle(display, x+int16(i), y+height-value, 1, value, white)
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
	for i := int16(0); i < 4; i++ {
		if i < int16(value) {
			tinydraw.FilledRectangle(display, x, y, width, height, white)
		} else {
			tinydraw.Rectangle(display, x, y, width, height, white)
		}
		x += width + gap
	}
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
	for i := int16(0); i < height; i++ {
		on := i >= height-value
		tinydraw.FilledRectangle(display, x, y+i*3, 2, 2, colorFor(on))
	}
}

func colorFor(on bool) color.RGBA {
	if on {
		return white
	}
	return black
}
