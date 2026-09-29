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

// Gauge ranges of the glance screens. The temperature gauge follows the sleep
// thermal scale, so its fill tracks apparent temperature while the printed
// number stays the measured one.
const (
	glanceCO2Min      float32 = 800
	glanceCO2Warn     float32 = 1300
	glanceCO2Max      float32 = 1400
	glanceApparentMin float32 = 24
	glanceApparentMax float32 = status.HotApparentC
	glanceHumidityMin float32 = 70
	glanceHumidityWet float32 = 80
	glanceHumidityMax float32 = 80
	glanceDewPointMin float32 = 16
	glanceDewPointWet float32 = 18
	glanceDewPointMax float32 = 23
)

// Geometry shared by the glance layouts.
const (
	glanceClockBand  int16 = 17
	glanceCellWidth  int16 = 40
	glanceCellGap    int16 = 4
	glanceRowTextX   int16 = 64
	glanceRowGaugeX  int16 = 91
	glanceRowGaugeW  int16 = 38
	glanceRowPitch   int16 = 11
	glanceNumberDrop int16 = 8
	glanceLabelDrop  int16 = 7
)

// glanceCell is one reading: the number to read up close and the gauge that
// carries it across the room.
type glanceCell struct {
	label           string
	text            string
	value, min, max float32
	mark            float32
}

// GlanceCentered draws the clock across the top and the three gauges along the
// bottom edge. A shorter bar height leaves more room below the clock.
//
//	            14:23
//	1200        29c         55%
//	|###|#|__|  |####|#|_|  |####|#|
func GlanceCentered(
	clockType font.FontType,
	barHeight int16,
) func(drivers.Displayer, *types.Readings) {
	return func(display drivers.Displayer, r *types.Readings) {
		if display == nil {
			return
		}

		widget.Clear(display)

		small := font.New(display, font.ProggySZ8)
		glanceClock(display, clockType, r)

		barY := glanceBarY(display, barHeight)

		x := int16(0)
		for _, cell := range glanceCells(r) {
			small.Print(x, barY-glanceNumberDrop, cell.text)
			glanceGaugeBordered(
				display,
				x,
				barY,
				glanceCellWidth,
				barHeight,
				cell,
			)
			x += glanceCellWidth + glanceCellGap
		}

		if err := display.Display(); err != nil {
			return
		}
	}
}

// GlanceLabels is GlanceCentered with the micro font, which frees the width
// each gauge needs for its own name.
//
//	            14:23
//	CO2 1200    T 29c       RH 55%
//	|###|#|__|  |####|#|_|  |####|#|
func GlanceLabels(
	clockType font.FontType,
	barHeight int16,
) func(drivers.Displayer, *types.Readings) {
	return func(display drivers.Displayer, r *types.Readings) {
		if display == nil {
			return
		}

		widget.Clear(display)

		tiny := font.New(display, font.TomThumb)
		glanceClock(display, clockType, r)

		barY := glanceBarY(display, barHeight)

		x := int16(0)
		for _, cell := range glanceCells(r) {
			tiny.Print(x, barY-glanceLabelDrop, cell.label+" "+cell.text)
			glanceGauge(display, x, barY, glanceCellWidth, barHeight, cell)
			x += glanceCellWidth + glanceCellGap
		}

		if err := display.Display(); err != nil {
			return
		}
	}
}

// GlanceLeft keeps the clock on the left half and stacks the readings on the
// right, so every number sits next to its own gauge.
//
//	         1200  |###|#|__|
//	14:23    29c   |####|#|_|
//	         55%   |####|#|
func GlanceLeft(
	clockType font.FontType,
	barHeight int16,
) func(drivers.Displayer, *types.Readings) {
	return func(display drivers.Displayer, r *types.Readings) {
		if display == nil {
			return
		}

		widget.Clear(display)

		var (
			small  = font.New(display, font.ProggySZ8)
			clock  = font.New(display, clockType)
			_, hgt = display.Size()
		)
		clock.Print(
			0,
			(hgt-clock.Height())/2,
			FormatTime(r.Time.Hour, r.Time.Minute),
		)

		y := int16(1)
		for _, cell := range glanceCells(r) {
			small.Print(glanceRowTextX, y, cell.text)
			glanceGaugeWithoutMark(
				display,
				glanceRowGaugeX,
				y+small.Height()/2-1,
				glanceRowGaugeW,
				barHeight,
				cell,
			)
			y += glanceRowPitch
		}

		if err := display.Display(); err != nil {
			return
		}
	}
}

func GlanceLeftSections(
	clockType font.FontType,
	barHeight int16,
	barWidth int16,
	barGap int16,
) func(drivers.Displayer, *types.Readings) {
	return func(display drivers.Displayer, r *types.Readings) {
		if display == nil {
			return
		}

		widget.Clear(display)

		var (
			small  = font.New(display, font.ProggySZ8)
			clock  = font.New(display, clockType)
			_, hgt = display.Size()
		)
		clock.Print(
			1,
			(hgt-clock.Height())/2,
			FormatTime(r.Time.Hour, r.Time.Minute),
		)

		y := int16(1)
		for _, cell := range glanceCells(r) {
			small.Print(glanceRowTextX, y, cell.text)
			glanceGaugeSections(
				display,
				glanceRowGaugeX,
				y+small.Height()/2-2,
				glanceRowGaugeW,
				barHeight,
				barWidth,
				barGap,
				cell,
			)
			y += glanceRowPitch
		}

		if err := display.Display(); err != nil {
			return
		}
	}
}

func glanceCells(r *types.Readings) [3]glanceCell {
	apparentC := status.ApparentTempC(r.Raw.Temperature, r.Raw.Humidity)
	humidity := float32(math.Round(float64(r.Raw.Humidity)))
	dewPoint := status.Dewpoint(float64(r.Raw.Temperature), float64(r.Raw.Humidity))

	return [3]glanceCell{
		{
			label: "CO2",
			text:  fmt.Sprintf("%d", r.Raw.CO2),
			value: float32(r.Raw.CO2),
			min:   glanceCO2Min,
			max:   glanceCO2Max,
			mark:  glanceCO2Warn,
		},
		{
			label: "T",
			text:  fmt.Sprintf("%.0fc", math.Round(float64(r.Raw.Temperature))),
			value: apparentC,
			min:   glanceApparentMin,
			max:   glanceApparentMax,
			mark:  status.WarmApparentC,
		},
		// {
		// 	label: "RH",
		// 	text:  fmt.Sprintf("%.0f%%", humidity),
		// 	value: humidity,
		// 	min:   glanceHumidityMin,
		// 	max:   glanceHumidityMax,
		// 	mark:  glanceHumidityWet,
		// },
		{
			label: "RH",
			text:  fmt.Sprintf("%.0f%%", humidity),
			value: float32(dewPoint),
			min:   glanceDewPointMin,
			max:   glanceDewPointWet,
			mark:  glanceDewPointMax,
		},
	}
}

func glanceClock(
	display drivers.Displayer,
	clockType font.FontType,
	r *types.Readings,
) {
	clock := font.New(display, clockType)
	layout.Center(
		display,
		clock,
		(glanceClockBand-clock.Height())/2,
		FormatTime(r.Time.Hour, r.Time.Minute),
	)
}

func glanceBarY(display drivers.Displayer, barHeight int16) int16 {
	_, height := display.Size()
	return height - barHeight
}

func glanceGauge(
	display drivers.Displayer,
	x, y, width, height int16,
	cell glanceCell,
) {
	widget.Gauge(
		display,
		x,
		y,
		width,
		height,
		cell.value,
		cell.min,
		cell.max,
		true,
	)
	widget.GaugeMark(
		display,
		x,
		y,
		width,
		height,
		cell.mark,
		cell.min,
		cell.max,
	)
}

func glanceGaugeWithoutMark(
	display drivers.Displayer,
	x, y, width, height int16,
	cell glanceCell,
) {
	widget.Gauge(
		display,
		x,
		y,
		width,
		height,
		cell.value,
		cell.min,
		cell.max,
		false,
	)
}

func glanceGaugeSections(
	display drivers.Displayer,
	x, y, width,
	barHeight int16,
	barWidth int16,
	barGap int16,
	cell glanceCell,
) {
	widget.SquareBar(
		display,
		x,
		y,
		width,
		barHeight,
		barWidth,
		barGap,
		cell.value,
		cell.min,
		cell.mark,
		cell.max,
	)
}

func glanceGaugeBordered(
	display drivers.Displayer,
	x, y, width, height int16,
	cell glanceCell,
) {
	widget.GaugeBordered(
		display,
		x,
		y,
		width,
		height,
		cell.value,
		cell.min,
		cell.max,
	)
	widget.GaugeMark(
		display,
		x,
		y,
		width,
		height,
		cell.mark,
		cell.min,
		cell.max,
	)
}
