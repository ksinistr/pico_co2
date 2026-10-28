package display

import (
	"fmt"
	"pico_co2/internal/clockedit"
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
func FormatEditTime(field clockedit.EditField, hour, minute int) string {
	switch field {
	case clockedit.FieldHour:
		return fmt.Sprintf("[%d]:%02d", hour, minute)
	case clockedit.FieldMinute:
		return fmt.Sprintf("%d:[%02d]", hour, minute)
	case clockedit.FieldSave:
		return "[SAVE] EXIT"
	case clockedit.FieldCancel:
		return "SAVE [EXIT]"
	default:
		return FormatTime(hour, minute)
	}
}

// RenderTime draws the main status screen and the clock editor.
//
//	H O .       26  55             C O O
//	              14:23
//	             T O .
func RenderTime(display drivers.Displayer, r *types.ViewState) {
	if display == nil {
		return
	}

	widget.Clear(display)
	renderTimeHeader(display, r)

	if r.ClockEdit.Active {
		text := FormatEditTime(r.ClockEdit.Field, r.ClockEdit.Hour, r.ClockEdit.Minute)
		layout.Center(display, font.New(display, font.FreemonoRegular9), 10, text)
	} else {
		layout.Center(display, font.New(display, font.FreemonoRegular18), 10, FormatClock(r.Time))
	}

	if err := display.Display(); err != nil {
		return
	}
}

func renderTimeHeader(display drivers.Displayer, r *types.ViewState) {
	small := font.New(display, font.ProggySZ8)
	width, _ := display.Size()

	if climateReady(r.Readings) {
		heatIndex := status.GetHeatIndex(float32(r.Readings.Temperature.Value), float32(r.Readings.Humidity.Value))
		widget.TwoSideBar(display, small, 0, 1, int16(heatIndex), "H", 0, 2)
	}

	temperature := readingText(r.Readings.Temperature, "")
	small.Print(width/2-small.Width(temperature)-3, 1, temperature)
	small.Print(width/2+2, 1, readingText(r.Readings.Humidity, ""))

	if r.Readings.CO2.State != types.ReadingReady {
		small.Print(97, 1, readingText(r.Readings.CO2, ""))
		return
	}

	var co2Status int16

	switch {
	case r.Readings.CO2.Value < 800:
		co2Status = 0
	case r.Readings.CO2.Value < 1000:
		co2Status = 1
	default:
		co2Status = 2
	}

	widget.TwoSideBar(display, small, 97, 1, co2Status, "C", 0, 2)
}
