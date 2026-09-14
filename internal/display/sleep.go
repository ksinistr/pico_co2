package display

import (
	"fmt"
	"math"
	"pico_co2/internal/display/font"
	"pico_co2/internal/types"
)

// RenderSleep shows bedroom sleep conditions.
//
// Status:
// OK means temperature/moisture/CO2 are acceptable for sleep.
// VENT means CO2 is high, open the window or increase ventilation.
// VENT? means ventilation is borderline.
// HOT means the room is too warm for sleep.
// HOT? means the room is still cooling.
// WET means the air is too moist, use AC/dehumidifier before sleep.
// WET? means moisture is borderline.
//
// DP is dew point: higher DP feels more sticky and means cooling alone may show 100% RH.
// DP <= 18C is good, 18-20C is borderline, >20C is wet, >22C is bad for sleep.
// AH is absolute humidity: grams of water per m3, useful for comparing nights.
// AH <= 15 is good, 15-17 is borderline, >17 is humid, >19 is bad for sleep.
func RenderSleep(renderer Renderer, r *types.Readings) {
	if renderer == nil {
		return
	}

	renderer.Clear()

	var (
		statusText = SleepStatus(
			r.Raw.CO2,
			r.Raw.Temperature,
			r.Calculated.DewPointC,
			r.Calculated.AbsoluteHumidityGM3,
		)
		dpText   = fmt.Sprintf("DP %.0fC", math.Round(float64(r.Calculated.DewPointC)))
		tempText = fmt.Sprintf("T%.0fC", math.Round(float64(r.Raw.Temperature)))
		ahText   = fmt.Sprintf("AH%.0f", math.Round(float64(r.Calculated.AbsoluteHumidityGM3)))
		co2Text  = fmt.Sprintf("C%d", r.Raw.CO2)
		sf       = renderer.GetFont(font.ProggySZ8)
	)

	width, _ := renderer.Size()
	sf.Print(0, 0, statusText)
	sf.Print(width-sf.CalcWidth(dpText), 0, dpText)

	y := int16(16)
	sf.Print(0, y, tempText)
	xAH := sf.CalcWidth(tempText) + (width-sf.CalcWidth(tempText)-sf.CalcWidth(ahText)-sf.CalcWidth(co2Text))/2
	sf.Print(xAH, y, ahText)
	sf.Print(width-sf.CalcWidth(co2Text), y, co2Text)

	renderer.Display()
}

func RenderSleepAdvice(renderer Renderer, r *types.Readings) {
	if renderer == nil {
		return
	}

	renderer.Clear()

	statusText := SleepStatus(
		r.Raw.CO2,
		r.Raw.Temperature,
		r.Calculated.DewPointC,
		r.Calculated.AbsoluteHumidityGM3,
	)
	line1, line2 := SleepAdvice(statusText)
	sf := renderer.GetFont(font.ProggySZ8)

	sf.Print(0, 0, statusText)
	sf.Print(0, 11, line1)
	sf.Print(0, 22, line2)

	renderer.Display()
}

func SleepStatus(co2 uint16, tempC, dewPointC, absoluteHumidityGM3 float32) string {
	switch {
	case co2 > 1200:
		return "VENT"
	case tempC >= 28:
		return "HOT"
	case dewPointC > 20 || absoluteHumidityGM3 > 17:
		return "WET"
	case co2 > 1000:
		return "VENT?"
	case tempC >= 26.5:
		return "HOT?"
	case dewPointC > 18 || absoluteHumidityGM3 > 15:
		return "WET?"
	default:
		return "OK"
	}
}

func SleepAdvice(statusText string) (string, string) {
	switch statusText {
	case "VENT":
		return "OPEN WINDOW", "NOW"
	case "VENT?":
		return "OPEN SMALL GAP", "WATCH CO2"
	case "HOT":
		return "AC COOL 25C", "ECO OFF"
	case "HOT?":
		return "WAIT COOLING", "ECO OFF IF SLOW"
	case "WET":
		return "CLOSE + DRY", "AC/DEHUMID"
	case "WET?":
		return "CLOSE WINDOW", "COOL TO 26C"
	default:
		return "KEEP SETTINGS", "SLEEP"
	}
}
