package widget

import (
	"image/color"

	"pico_co2/pkg/font"
	"pico_co2/pkg/miniplot"
	"pico_co2/pkg/sparkline"
	"tinygo.org/x/drivers"
	"tinygo.org/x/tinydraw"
)

var (
	white = color.RGBA{255, 255, 255, 255}
	black = color.RGBA{0, 0, 0, 255}
)

type TrendDirection uint8

const (
	TrendStable TrendDirection = iota
	TrendRising
	TrendFalling
	TrendUnknown
)

func Clear(display drivers.Displayer) {
	if display == nil {
		return
	}
	width, height := display.Size()
	tinydraw.FilledRectangle(display, 0, 0, width, height, black)
}

func TwoSideBar(display drivers.Displayer, face font.Face, x, y, value int16, label string, left, right int16) int16 {
	if display == nil {
		return x
	}
	const radius, spacing int16 = 3, 5
	barY := y + radius + 1
	for i := int16(0); i < left; i++ {
		barX := x + radius + i*(2*radius+spacing)
		r := int16(1)
		if value < 0 && i >= 3+value {
			r = radius
		}
		tinydraw.FilledCircle(display, barX, barY, r, white)
	}
	x += left*(2*radius+spacing) + radius
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
	return x + right*(2*radius+spacing) + radius
}

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

func ScaleX(width int16, value, min, max float32) int16 {
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}
	return int16(float32(width) * (value - min) / (max - min))
}

func ThermalScale(display drivers.Displayer, y, height int16, value, min, warm, hot, max float32) {
	if display == nil {
		return
	}
	width, _ := display.Size()
	origin := ScaleX(width, warm, min, max)
	tinydraw.FilledRectangle(display, 0, y+height-1, origin, 1, white)
	tinydraw.FilledRectangle(display, origin, y-2, 1, height+2, white)
	valueX := ScaleX(width, value, min, max)
	if valueX <= origin {
		return
	}
	tinydraw.FilledRectangle(display, origin, y, valueX-origin, height, white)
	hotX := ScaleX(width, hot, min, max)
	tinydraw.FilledRectangle(display, hotX-1, y, 3, height, black)
	tinydraw.FilledRectangle(display, hotX, y, 1, height, white)
}

func Sparkline(display drivers.Displayer, x, y, width, height int16, data []int16) {
	if len(data) == 0 {
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

func Plot(display drivers.Displayer, face font.Face, data []int16, title string) error {
	width, height := display.Size()
	plot, err := miniplot.NewMiniPlot(display, face.Raw(), width, height, white)
	if err != nil {
		return err
	}
	plot.DrawLineChart(data, title)
	return nil
}

func colorFor(on bool) color.RGBA {
	if on {
		return white
	}
	return black
}
