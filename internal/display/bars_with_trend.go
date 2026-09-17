package display

import (
	"fmt"
	"math"

	"pico_co2/internal/display/font"
	"pico_co2/internal/types"
	"pico_co2/internal/types/status"
)

// RenderBarsWithTrend is the three number layout with units, a CO2 trend arrow
// in the header and CO2 as the largest value. The dots next to T follow the
// sleep thermal zone, which unlike the heat index keeps moving below 27C.
func RenderBarsWithTrend(renderer Renderer, r *types.Readings) {
	if renderer == nil {
		return
	}

	renderer.Clear()

	var (
		sf = renderer.GetFont(font.ProggySZ8)
		mf = renderer.GetFont(font.FreemonoRegular9)
		lf = renderer.GetFont(font.FreemonoRegular12)
	)

	width, _ := renderer.Size()
	apparentC := status.ApparentTempC(r.Raw.Temperature, r.Raw.Humidity)

	renderer.DrawTwoSideBar(0, 0, int16(status.ToSleepThermal(apparentC)), "T", 0, 2)
	sf.Print(98, 1, "CO2")
	DrawTrend(renderer, 122, 5, r.Calculated.CO2Trend)

	timeStr := FormatTime(r.Time.Hour, r.Time.Minute)
	sf.Print((width-sf.CalcWidth(timeStr))/2, 1, timeStr)

	tempStr := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Temperature)))
	humStr := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Humidity)))
	co2Str := fmt.Sprintf("%d", r.Raw.CO2)

	tempWidth := mf.CalcWidth(tempStr) + sf.CalcWidth("c")
	humWidth := mf.CalcWidth(humStr) + sf.CalcWidth("%")
	co2Width := lf.CalcWidth(co2Str)
	gap := (width - tempWidth - humWidth - co2Width) / 2

	y := int16(17)

	mf.Print(0, y, tempStr)
	sf.Print(mf.CalcWidth(tempStr)+1, y+5, "c")

	xHum := tempWidth + gap
	mf.Print(xHum, y, humStr)
	sf.Print(xHum+mf.CalcWidth(humStr)+1, y+5, "%")

	lf.Print(width-co2Width, y-2, co2Str)

	renderer.Display()
}
