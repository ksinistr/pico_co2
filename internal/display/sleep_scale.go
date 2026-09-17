package display

import (
	"fmt"
	"math"

	"pico_co2/internal/display/font"
	"pico_co2/internal/types"
	"pico_co2/internal/types/status"
)

// RenderSleepScale shows the two values an air conditioned bedroom is steered
// by - CO2 and temperature - at the same size, with the sleep thermal scale
// along the bottom edge.
func RenderSleepScale(renderer Renderer, r *types.Readings) {
	if renderer == nil {
		return
	}

	renderer.Clear()

	var (
		sf = renderer.GetFont(font.ProggySZ8)
		lf = renderer.GetFont(font.FreemonoRegular12)
	)

	width, _ := renderer.Size()
	apparentC := status.ApparentTempC(r.Raw.Temperature, r.Raw.Humidity)

	timeStr := FormatTime(r.Time.Hour, r.Time.Minute)
	humStr := fmt.Sprintf("%.0f%%", math.Round(float64(r.Raw.Humidity)))
	wordStr := status.ToSleepThermal(apparentC).String()

	sf.Print(0, 0, timeStr)
	sf.Print((width-sf.CalcWidth(humStr))/2, 0, humStr)
	sf.Print(width-sf.CalcWidth(wordStr), 0, wordStr)

	co2Str := fmt.Sprintf("%d", r.Raw.CO2)
	tempStr := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Temperature)))
	y := int16(8)

	lf.Print(0, y, co2Str)
	DrawTrend(renderer, lf.CalcWidth(co2Str)+7, y+8, r.Calculated.CO2Trend)

	xTemp := width - lf.CalcWidth(tempStr) - sf.CalcWidth("c") - 1
	lf.Print(xTemp, y, tempStr)
	sf.Print(width-sf.CalcWidth("c"), y+9, "c")

	DrawThermalScale(renderer, 27, 5, apparentC)

	renderer.Display()
}
