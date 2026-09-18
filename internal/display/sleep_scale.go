package display

import (
	"fmt"
	"math"
	"pico_co2/internal/types"
	"pico_co2/internal/types/status"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"

	"tinygo.org/x/drivers"
)

// RenderSleepScale shows the two values an air conditioned bedroom is steered
// by - CO2 and temperature - at the same size, with the sleep thermal scale
// along the bottom edge.
//
//	14:23          55%              WARM
//	1200 /\                         26 c
//	______________|###########|#|_______
func RenderSleepScale(display drivers.Displayer, r *types.Readings) {
	if display == nil {
		return
	}

	widget.Clear(display)

	var (
		sf = font.New(display, font.ProggySZ8)
		lf = font.New(display, font.FreemonoRegular12)
	)

	width, _ := display.Size()
	apparentC := status.ApparentTempC(r.Raw.Temperature, r.Raw.Humidity)

	timeStr := FormatTime(r.Time.Hour, r.Time.Minute)
	humStr := fmt.Sprintf("%.0f%%", math.Round(float64(r.Raw.Humidity)))
	wordStr := status.ToSleepThermal(apparentC).String()

	layout.Three(display, sf, 0, timeStr, humStr, wordStr)

	co2Str := fmt.Sprintf("%d", r.Raw.CO2)
	tempStr := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Temperature)))
	y := int16(8)

	lf.Print(0, y, co2Str)
	widget.Trend(display, lf.Width(co2Str)+7, y+8, trendDirection(r.Calculated.CO2Trend))

	xTemp := width - lf.Width(tempStr) - sf.Width("c") - 1
	lf.Print(xTemp, y, tempStr)
	sf.Print(width-sf.Width("c"), y+9, "c")

	widget.ThermalScale(display, 27, 3, apparentC, widget.ThermalRange{
		Min:  24,
		Warm: status.WarmApparentC,
		Hot:  status.HotApparentC,
		Max:  36,
	})

	if err := display.Display(); err != nil {
		return
	}
}
