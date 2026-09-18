// Package miniplot draws a compact auto-scrolling line chart.
package miniplot

import (
	"errors"
	"fmt"
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
)

// MiniPlot holds the display and text metrics used by a line chart.
type MiniPlot struct {
	display       drivers.Displayer
	font          tinyfont.Fonter
	fontWidth     int16 // Width of a single character in the font
	fontHeight    int16 // Height of a single character in the font
	DisplayWidth  int16
	DisplayHeight int16
	Color         color.RGBA
	StartX        int16 // X position to start drawing the plot
	StartY        int16 // Y position to start drawing the plot
}

// NewMiniPlot constructs a chart for the supplied display dimensions.
func NewMiniPlot(
	display drivers.Displayer,
	font tinyfont.Fonter,
	displayWidth int16,
	displayHeight int16,
	c color.RGBA,
) (*MiniPlot, error) {
	if displayWidth <= 0 || displayHeight <= 0 {
		return nil, errors.New("display dimensions must be greater than zero")
	}

	if display == nil {
		return nil, errors.New("display cannot be nil")
	}

	if font == nil {
		return nil, errors.New("font cannot be nil")
	}

	return &MiniPlot{
		display:       display,
		font:          font,
		fontWidth:     int16(tinyfont.GetGlyph(font, '0').Info().Width),
		fontHeight:    int16(tinyfont.GetGlyph(font, '0').Info().Height),
		DisplayWidth:  displayWidth,
		DisplayHeight: displayHeight,
		Color:         c,
		StartX:        20,
		StartY:        22,
	}, nil
}

// DrawLineChart draws a title, extrema, time axis, and right-aligned samples.
//
//	title  max
//	       /\    /\
//	min __/  \__/  \____
//	-8h      -1h    now
func (mp *MiniPlot) DrawLineChart(
	data []int16,
	title string,
) error {
	if len(data) == 0 {
		return nil
	}

	// Cut slice to fit display width, only keep the last N samples
	if len(data) > int(mp.DisplayWidth-mp.StartX-2) {
		data = data[len(data)-int(mp.DisplayWidth-mp.StartX-2):]
	}

	minVal := data[0]

	maxVal := data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}

		if v > maxVal {
			maxVal = v
		}
	}

	if err := tinydraw.FilledRectangle(
		mp.display,
		0,
		0,
		mp.DisplayWidth,
		mp.DisplayHeight,
		color.RGBA{0, 0, 0, 255},
	); err != nil {
		return fmt.Errorf("clear plot: %w", err)
	}

	mp.drawAxis(maxVal, minVal, title)
	mp.drawData(data, minVal, maxVal)

	if err := mp.display.Display(); err != nil {
		return fmt.Errorf("display plot: %w", err)
	}

	return nil
}

func (mp *MiniPlot) drawText(x, y int16, text string) {
	tinyfont.WriteLine(mp.display, mp.font, x, y, text, mp.Color)
}

func (mp *MiniPlot) drawAxis(
	maxValue, minValue int16,
	title string,
) {
	startX := mp.StartX // Start X position for the axis
	startY := mp.StartY // Start Y position for the axis

	tinydraw.Line(mp.display, startX, startY, startX, 0, mp.Color)
	tinydraw.Line(mp.display, startX, startY, mp.DisplayWidth-1, startY, mp.Color)
	mp.drawText(1, startY+mp.fontHeight, title)

	label := formatValue(minValue)
	mp.drawText(1, startY, label)

	label = formatValue(maxValue)
	mp.drawText(1, mp.fontHeight, label)

	text := "0h"
	textWidth := int16(len(text)) * mp.fontWidth
	xPos := mp.DisplayWidth - textWidth
	mp.drawText(xPos, startY+mp.fontHeight, text)

	text = "-1h"
	textWidth = int16(len(text)) * mp.fontWidth
	halfTextWidth := textWidth / 2
	xPos = mp.DisplayWidth - int16(60) - halfTextWidth
	mp.drawText(xPos, startY+mp.fontHeight, text)
}

// drawData draws an auto-scrolling line chart that grows from the right edge.
func (mp *MiniPlot) drawData(samples []int16, minV, maxV int16) {
	var (
		baseY  = int16(22)           // the horizontal axis pixel row
		rightX = mp.DisplayWidth - 2 // right-most drawable column
		topY   = int16(1)            // graph top (screen row 1)
	)

	if len(samples) == 0 {
		return
	}

	rangeVal := float64(maxV - minV)
	if rangeVal == 0 {
		rangeVal = 1
	}

	pixelsPerUnit := float64(baseY-topY) / rangeVal

	pixelY := func(v int16) int16 {
		return int16(float64(maxV-v) * pixelsPerUnit)
	}

	n := len(samples)
	for i := 1; i < n; i++ {
		x1 := rightX - int16(i-1)
		x2 := rightX - int16(i)
		y1 := pixelY(samples[n-i])
		y2 := pixelY(samples[n-i-1])

		tinydraw.Line(mp.display, x1, y1, x2, y2, mp.Color)
	}
}

func formatValue(value int16) string {
	if value >= 1000 {
		return fmt.Sprintf("%dk", value/1000)
	}

	return fmt.Sprintf("%d", value)
}
