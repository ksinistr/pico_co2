package display

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinydraw"

	"pico_co2/internal/display/bar"
	"pico_co2/internal/display/font"
	"pico_co2/pkg/miniplot"
	"pico_co2/pkg/sparkline"
)

// SSD1306Adapter wraps an ssd1306 display
type SSD1306Adapter struct {
	dev   drivers.Displayer
	white color.RGBA
	black color.RGBA
	fonts *font.FontRegistry
}

func NewSSD1306Adapter(dev drivers.Displayer) *SSD1306Adapter {
	white := color.RGBA{255, 255, 255, 255}
	return &SSD1306Adapter{
		dev:   dev,
		white: white,
		black: color.RGBA{0, 0, 0, 255},
		fonts: font.NewFontRegistry(dev, white),
	}
}

func (v *SSD1306Adapter) SetPixel(x, y int16, c color.RGBA) {
	v.dev.SetPixel(x, y, c)
}

func (v *SSD1306Adapter) Size() (int16, int16) {
	return v.dev.Size()
}

func (v *SSD1306Adapter) Clear() {
	width, height := v.dev.Size()
	for x := range width {
		for y := range height {
			v.dev.SetPixel(x, y, v.black)
		}
	}
}

func (v *SSD1306Adapter) Display() error {
	v.dev.Display()
	return nil
}

// NEW: Unified font management methods
func (v *SSD1306Adapter) GetFont(fontType font.FontType) font.FontPrinter {
	if v == nil || v.fonts == nil {
		return nil
	}
	return v.fonts.GetFont(fontType)
}

func (v *SSD1306Adapter) DrawText(fontType font.FontType, x, y int16, text string) {
	if v == nil || v.fonts == nil {
		return
	}
	font := v.fonts.GetFont(fontType)
	if font != nil {
		font.Print(x, y, text)
	}
}

func (v *SSD1306Adapter) CalcTextWidth(fontType font.FontType, text string) int16 {
	if v == nil || v.fonts == nil {
		return 0
	}
	font := v.fonts.GetFont(fontType)
	if font != nil {
		return font.CalcWidth(text)
	}
	return 0
}

// Legacy font methods for backward compatibility
func (v *SSD1306Adapter) DrawXLargeText(x, y int16, text string) {
	v.DrawText(font.FreemonoBold18, x, y, text)
}

func (v *SSD1306Adapter) CalcXLargeTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreemonoBold18, text)
}

func (v *SSD1306Adapter) DrawLargeText(x, y int16, text string) {
	v.DrawText(font.FreemonoRegular12, x, y, text)
}

func (v *SSD1306Adapter) DrawLargeBoldText(x, y int16, text string) {
	v.DrawText(font.FreemonoBold12, x, y, text)
}

func (v *SSD1306Adapter) DrawLargeSansText(x, y int16, text string) {
	v.DrawText(font.FreesansRegular12, x, y, text)
}

func (v *SSD1306Adapter) CalcLargeTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreemonoRegular12, text)
}

func (v *SSD1306Adapter) CalcLargeBoldTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreemonoBold12, text)
}

func (v *SSD1306Adapter) CalcLargeSansTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreesansRegular12, text)
}

func (v *SSD1306Adapter) DrawSmallText(x, y int16, text string) {
	v.DrawText(font.ProggySZ8, x, y, text)
}

func (v *SSD1306Adapter) CalcSmallTextWidth(text string) int16 {
	return v.CalcTextWidth(font.ProggySZ8, text)
}

func (v *SSD1306Adapter) DrawLongText(x, y int16, text string) {
	if v == nil || v.fonts == nil {
		return
	}

	font := v.fonts.GetFont(font.ProggySZ8)
	if font == nil {
		return
	}

	width, height := v.Size()

	// Wrap text to fit the display width, accounting for the initial x offset.
	lines := wrapText(text, font, width-x)

	fontHeight := font.Height()
	lineHeight := fontHeight + 1 // Font height + 1px spacing

	// Calculate how many lines can be drawn in the available vertical space.
	availableHeight := height - y
	if availableHeight < 0 {
		availableHeight = 0
	}
	maxLines := int(availableHeight / lineHeight)

	for i, line := range lines {
		if i >= maxLines {
			break
		}
		// Y-coordinate for the baseline of the current line.
		drawY := y + fontHeight + (int16(i) * lineHeight)
		font.Print(x, drawY, line)
	}
}

func (v *SSD1306Adapter) DrawPlot(data []int16, title string) {
	if v == nil || v.fonts == nil {
		return
	}

	font := v.fonts.GetFont(font.ProggySZ8)
	if font == nil {
		return
	}

	width, height := v.Size()
	miniPlot, err := miniplot.NewMiniPlot(
		v,
		font.GetFont(),
		width,
		height,
		v.white,
	)
	if err != nil {
		println("Error creating CO2 graph:", err.Error())
		return
	}

	miniPlot.DrawLineChart(data, "CO2")
}

func (v *SSD1306Adapter) DrawTwoSideBar(
	x, y int16,
	value int16,
	label string,
	leftCount int16,
	rightCount int16,
) int16 {
	if v == nil || v.fonts == nil {
		return y
	}

	var (
		radius     int16 = 3
		barSpacing int16 = 5
	)

	font := v.fonts.GetFont(font.ProggySZ8)
	if font == nil {
		return y
	}

	hbar := bar.NewTwoSideBar(
		v,
		radius,
		barSpacing,
		leftCount,
		rightCount,
		font.GetFont(),
		v.white,
	)

	return hbar.Draw(x, y, value, label)
}

func (v *SSD1306Adapter) DrawSparkline(
	x, y int16,
	data []int16,
	width int16,
	height int16,
) {
	if len(data) == 0 {
		return
	}

	sp := sparkline.NewSparkline(
		int(height),
	)

	if len(data) > int(width) {
		data = data[len(data)-int(width):]
	}

	result := sp.Process(data)

	y = y + height

	for i, barHeight := range result {
		topY := y - barHeight

		tinydraw.FilledRectangle(v, x+int16(i), topY, barWidth, barHeight, v.white)
	}
}

func (v *SSD1306Adapter) colorFor(on bool) color.RGBA {
	if on {
		return v.white
	}
	return v.black
}

func (v *SSD1306Adapter) FillRect(x, y, w, h int16, on bool) {
	tinydraw.FilledRectangle(v, x, y, w, h, v.colorFor(on))
}

func (v *SSD1306Adapter) FillTriangle(x0, y0, x1, y1, x2, y2 int16, on bool) {
	tinydraw.FilledTriangle(v, x0, y0, x1, y1, x2, y2, v.colorFor(on))
}

func (v *SSD1306Adapter) DrawSquareBar(x, y int16, value uint8) {
	if value == 0 {
		return
	}

	squareWidth := int16((128 - 6*3) / 4)
	squareHeight := int16(9)
	totalCount := 4
	filledCount := int16(value)

	for range totalCount {
		if filledCount > 0 {
			filledCount--
			tinydraw.FilledRectangle(v, x, y, squareWidth, squareHeight, v.white)
		} else {
			tinydraw.Rectangle(v, x, y, squareWidth, squareHeight, v.white)
		}
		x = x + squareWidth + 6
	}
}

