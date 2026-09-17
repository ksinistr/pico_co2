package display

import (
	"fmt"

	"pico_co2/internal/types"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"
	"tinygo.org/x/drivers"
)

// RenderSparklineCO2 draws eight hours of CO2 history against 1000 ppm.
//
//	8h CO2 700-1300                    42%
//	      #       #
//	  # # # ##  # ###
func RenderSparklineCO2(display drivers.Displayer, r *types.Readings) {
	data := r.History.CO2.Contiguous()
	title := "CO2"
	baseline := int16(1000)

	renderSparkline(display, title, data, baseline)
}

// RenderSparklineT draws eight hours of temperature history against 27 C.
//
//	8h T 23-29                         35%
//	      #       #
//	  # # # ##  # ###
func RenderSparklineT(display drivers.Displayer, r *types.Readings) {
	data := r.History.Temperature.Contiguous()
	title := "T"
	baseline := int16(27)

	renderSparkline(display, title, data, baseline)
}

// RenderSparklineRH draws eight hours of humidity history against 45% RH.
//
//	8h RH 40-65                        58%
//	      #       #
//	  # # # ##  # ###
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
