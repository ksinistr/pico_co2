package display

import (
	"fmt"
	"math"
	"pico_co2/internal/display/font"
	"pico_co2/internal/types"
	"pico_co2/internal/types/status"
)

func RenderBarsWithLargeNums(renderer Renderer, r *types.Readings) {
	if renderer == nil {
		return
	}

	renderer.Clear()

	var (
		y         int16 = 1
		x         int16
		co2status int16
		lf        = renderer.GetFont(font.FreemonoRegular9)
		sf        = renderer.GetFont(font.ProggySZ8)
	)

	width, _ := renderer.Size()

	// First line
	heatIndex := status.GetHeatIndex(r.Raw.Temperature, r.Raw.Humidity)
	x = renderer.DrawTwoSideBar(x, y, int16(heatIndex), "T", 0, 2)

	// https://backend.orbit.dtu.dk/ws/portalfiles/portal/348932926/1-s2.0-S0360132323011459-main_1_.pdf
	switch {
	case r.Raw.CO2 < 800:
		co2status = 0
	case r.Raw.CO2 < 1000:
		co2status = 1
	default:
		co2status = 2
	}
	x = 96
	renderer.DrawTwoSideBar(x, y, co2status, "C", 0, 2)

	lineStr := FormatTime(r.Time.Hour, r.Time.Minute)
	xLine := (width - sf.CalcWidth(lineStr)) / 2
	sf.Print(xLine, y, lineStr)

	// second line
	x = 0
	y = 16
	tempStr := fmt.Sprintf(
		"%.0f",
		math.Round(float64(r.Raw.Temperature)),
	)
	lf.Print(x, y, tempStr)
	humStr := fmt.Sprintf(
		"%.0f",
		math.Round(float64(r.Raw.Humidity)),
	)
	co2str := fmt.Sprintf("%d", r.Raw.CO2)
	xHum := lf.CalcWidth(
		tempStr,
	) + (width-lf.CalcWidth(tempStr)-lf.CalcWidth(humStr)-lf.CalcWidth(co2str))/2
	lf.Print(xHum, y, humStr)

	xCO2 := width - lf.CalcWidth(co2str)
	lf.Print(xCO2, y, co2str)

	renderer.Display()
}
