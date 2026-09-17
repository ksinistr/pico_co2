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

// FormatTime returns the normal display string for a clock time.
func FormatTime(hour, minute int) string {
	return fmt.Sprintf("%d:%02d", hour, minute)
}

// FormatEditTime returns the display string for clock edit mode.
// The selected field is indicated by square brackets for numeric fields
// or by the action label for save/cancel fields.
func FormatEditTime(field int, hour, minute int) string {
	switch field {
	case types.EditFieldHour:
		return fmt.Sprintf("[%d]:%02d", hour, minute)
	case types.EditFieldMinute:
		return fmt.Sprintf("%d:[%02d]", hour, minute)
	case types.EditFieldSave:
		return "[SAVE] EXIT"
	case types.EditFieldCancel:
		return "SAVE [EXIT]"
	default:
		return FormatTime(hour, minute)
	}
}

func RenderTime(display drivers.Displayer, r *types.Readings) {
	if display == nil {
		return
	}

	widget.Clear(display)

	var (
		y         int16 = 1
		x         int16
		co2status int16
		lf        = font.New(display, font.FreemonoRegular18)
		mf        = font.New(display, font.FreemonoRegular9)
		sf        = font.New(display, font.ProggySZ8)
	)

	width, _ := display.Size()

	// First line

	heatIndex := status.GetHeatIndex(r.Raw.Temperature, r.Raw.Humidity)
	x = widget.TwoSideBar(display, sf, x, y, int16(heatIndex), "H", 0, 2)

	temp := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Temperature)))
	hum := fmt.Sprintf("%.0f", math.Round(float64(r.Raw.Humidity)))
	x = width/2 - sf.Width(temp) - 2 - 1
	sf.Print(x, y, temp)
	x = width/2 + 2
	sf.Print(x, y, hum)

	// https://backend.orbit.dtu.dk/ws/portalfiles/portal/348932926/1-s2.0-S0360132323011459-main_1_.pdf
	switch {
	case r.Raw.CO2 < 800:
		co2status = 0
	case r.Raw.CO2 < 1000:
		co2status = 1
	default:
		co2status = 2
	}
	x = 97
	widget.TwoSideBar(display, sf, x, y, co2status, "C", 0, 2)

	// second line
	y = 10
	var lineStr string
	if r.ClockEdit.Active {
		lineStr = FormatEditTime(r.ClockEdit.Field, r.ClockEdit.Hour, r.ClockEdit.Minute)
		layout.Center(display, mf, y, lineStr)
	} else {
		lineStr = FormatTime(r.Time.Hour, r.Time.Minute)
		layout.Center(display, lf, y, lineStr)
	}

	display.Display()
}
