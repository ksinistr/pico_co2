package display

import (
	"fmt"
	"pico_co2/internal/types"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"
	"tinygo.org/x/drivers"
)

func RenderSparklineCO2(display drivers.Displayer, r *types.Readings) {
	data := r.History.CO2.Contiguous()
	title := "CO2"
	baseline := int16(1000)

	renderSparkline(display, title, data, baseline)
}

func RenderSparklineT(display drivers.Displayer, r *types.Readings) {
	data := r.History.Temperature.Contiguous()
	title := "T"
	baseline := int16(27)

	renderSparkline(display, title, data, baseline)
}

func RenderSparklineRH(display drivers.Displayer, r *types.Readings) {
	data := r.History.Humidity.Contiguous()
	title := "RH"
	baseline := int16(45)

	renderSparkline(display, title, data, baseline)
}

func renderSparkline(
	display drivers.Displayer,
	title string,
	data []int16,
	baseline int16,
) {
	if display == nil {
		return
	}

	widget.Clear(display)

	var (
		y  int16
		x  int16
		sf = font.New(display, font.ProggySZ8)
	)
	minV, maxV := minMaxInt16Slice(data)
	percentAbove := calcPercentAboveBaseline(data, baseline)

	titleStr := fmt.Sprintf("8h %s %d-%d", title, minV, maxV)
	sf.Print(0, 0, titleStr)

	sparklineTitle := fmt.Sprintf("%.0f%%", percentAbove)
	layout.Right(display, sf, 0, sparklineTitle)

	x = 0
	y = 11

	var (
		graphHeight int16 = 21
		graphWidth  int16 = 128
	)

	widget.Sparkline(display, x, y, graphWidth, graphHeight, data)
	display.Display()
}

func minMaxInt16Slice(data []int16) (minV int16, maxV int16) {
	if len(data) == 0 {
		return 0, 0
	}
	minV = data[0]
	maxV = data[0]
	for _, v := range data {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	return minV, maxV
}

// calculate percent above baseline in slice of int16
func calcPercentAboveBaseline(data []int16, baseline int16) float32 {
	if len(data) == 0 {
		return 0
	}
	countAbove := 0
	for _, v := range data {
		if v > baseline {
			countAbove++
		}
	}
	return float32(countAbove) / float32(len(data)) * 100.0
}
