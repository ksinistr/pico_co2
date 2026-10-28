package display

import (
	"strconv"

	"pico_co2/internal/types"
	"pico_co2/internal/types/status"
	"pico_co2/pkg/font"
	"pico_co2/pkg/widget"

	"tinygo.org/x/drivers"
)

func EasyRead(display drivers.Displayer, view *types.ViewState) {
	if display == nil {
		return
	}

	widget.Clear(display)

	var (
		xsmall        = font.New(display, font.ProggySZ8)
		small         = font.New(display, font.FreemonoRegular9)
		timeFont      = font.New(display, font.FreemonoRegular12)
		width, height = display.Size()
		apparentC     = float64(status.ApparentTempC(
			float32(view.Readings.Temperature.Value),
			float32(view.Readings.Humidity.Value),
		))
		hour      = view.Time.Hour
		minute    = view.Time.Minute
		hourStr   = strconv.Itoa(hour)
		minuteStr = strconv.Itoa(minute)
	)

	if len(hourStr) == 1 {
		hourStr = "0" + hourStr
	}

	if len(minuteStr) == 1 {
		minuteStr = "0" + minuteStr
	}

	var (
		hourMinuteString = hourStr + ":" + minuteStr
		// hourMinuteStringWidth = timeBold.Width(hourMinuteString)
		co2      = strconv.Itoa(int(view.Readings.CO2.Value))
		humidity = strconv.Itoa(int(view.Readings.Humidity.Value))
		temp     = strconv.Itoa(
			int(view.Readings.Temperature.Value),
		)
		co2Width = small.Width(co2)
		// humidityWidth = small.Width(humidity)
		apparentCStr = strconv.Itoa(int(apparentC))
		// temperatureWidth = small.Width(temp)
		apparentCWidth = small.Width(apparentCStr)
		bottomLine     = "T " + temp + " H " + humidity
	)

	// time
	timeFont.Print(0, 0, hourMinuteString)

	// co2
	small.Print(width-co2Width, 0, co2)

	xsmall.Print(0, height-xsmall.Height()-1, bottomLine)
	// // temperature
	// small.Print(0, height/2, temp)
	// // humidity
	// small.Print(temperatureWidth, height/2, humidity)
	// apparent temperature

	xsmall.Print(width-apparentCWidth-xsmall.Width("AT "), height-xsmall.Height()-1, "AT ")
	small.Print(width-apparentCWidth, height/2+4, apparentCStr)

	display.Display()
}
