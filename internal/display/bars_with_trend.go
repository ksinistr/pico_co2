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

// RenderBarsWithTrend is the three number layout with units, a CO2 trend arrow
// in the header and CO2 as the largest value. The dots next to T follow the
// sleep thermal zone, which unlike the heat index keeps moving below 27C.
//
//	T O .        14:23              CO2 /\
//	24 c             55 %             1200
func RenderBarsWithTrend(display drivers.Displayer, r *types.Readings) {
	if display == nil {
		return
	}

	widget.Clear(display)

	var (
		sf = font.New(display, font.ProggySZ8)
		mf = font.New(display, font.FreemonoRegular9)
		lf = font.New(display, font.FreemonoRegular12)
	)

	apparentC := status.ApparentTempC(r.Raw.Temperature, r.Raw.Humidity)

	widget.TwoSideBar(display, sf, 0, 0, int16(status.ToSleepThermal(apparentC)), "T", 0, 2)
	sf.Print(98, 1, "CO2")
	widget.Trend(display, 122, 5, trendDirection(r.Calculated.CO2Trend))

	timeStr := FormatTime(r.Time.Hour, r.Time.Minute)
	layout.Center(display, sf, 1, timeStr)

	tempStr := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Temperature)))
	humStr := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Humidity)))
	co2Str := fmt.Sprintf("%d", r.Raw.CO2)

	layout.Row(display, 17,
		layout.Cell{Face: mf, Text: tempStr, Unit: "c", UnitFace: sf},
		layout.Cell{Face: mf, Text: humStr, Unit: "%", UnitFace: sf},
		layout.Cell{Face: lf, Text: co2Str, YOffset: -2},
	)

	display.Display()
}
