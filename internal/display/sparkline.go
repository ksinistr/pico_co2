package display

import (
	"fmt"
	"pico_co2/internal/display/font"
	"pico_co2/internal/types"
)

func RenderSparklineCO2(renderer Renderer, r *types.Readings) {
	data := r.History.CO2.Contiguous()
	title := "CO2"
	baseline := int16(1000)

	renderSparkline(renderer, title, data, baseline)
}

func RenderSparklineT(renderer Renderer, r *types.Readings) {
	data := r.History.Temperature.Contiguous()
	title := "T"
	baseline := int16(27)

	renderSparkline(renderer, title, data, baseline)
}

func RenderSparklineRH(renderer Renderer, r *types.Readings) {
	data := r.History.Humidity.Contiguous()
	title := "RH"
	baseline := int16(45)

	renderSparkline(renderer, title, data, baseline)
}

func RenderSparklineHI(renderer Renderer, r *types.Readings) {
	data := r.History.HeatIndexTemp.Contiguous()
	title := "HI"
	baseline := int16(27)

	renderSparkline(renderer, title, data, baseline)
}

func RenderSparklineDP(renderer Renderer, r *types.Readings) {
	data := r.History.DewPoint.Contiguous()
	title := "DP"
	baseline := int16(18)

	renderSparkline(renderer, title, data, baseline)
}

func renderSparkline(
	renderer Renderer,
	title string,
	data []int16,
	baseline int16,
) {
	if renderer == nil {
		return
	}

	renderer.Clear()

	var (
		y  int16
		x  int16
		sf = renderer.GetFont(font.ProggySZ8)
	)
	minV, maxV := minMaxInt16Slice(data)
	percentAbove := calcPercentAboveBaseline(data, baseline)

	titleStr := fmt.Sprintf("8h %s %d-%d", title, minV, maxV)
	sf.Print(0, 0, titleStr)

	sparklineTitle := fmt.Sprintf("%.0f%%", percentAbove)
	width, _ := renderer.Size()
	sf.Print(width-sf.CalcWidth(sparklineTitle), 0, sparklineTitle)

	x = 0
	y = 11

	var (
		graphHeight int16 = 21
		graphWidth  int16 = 128
	)

	renderer.DrawSparkline(x, y, data, graphWidth, graphHeight)
	renderer.Display()
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
