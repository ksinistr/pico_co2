package display

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/nfnt/resize"
	"tinygo.org/x/tinydraw"

	"pico_co2/internal/display/bar"
	"pico_co2/internal/display/font"
	"pico_co2/pkg/miniplot"
	"pico_co2/pkg/sparkline"
)

type VirtualDisplay struct {
	buffer []color.RGBA
	width  int16
	height int16
	white  color.RGBA
	black  color.RGBA
	fonts  *font.FontRegistry
}

func NewVirtualDisplay(w, h int16) *VirtualDisplay {
	size := w * h
	white := color.RGBA{255, 255, 255, 255}
	v := &VirtualDisplay{
		buffer: make([]color.RGBA, size),
		width:  w,
		height: h,
		white:  white,
		black:  color.RGBA{0, 0, 0, 255},
	}
	v.fonts = font.NewFontRegistry(v, white)
	return v
}

func (v *VirtualDisplay) Clear() {
	for i := range v.buffer {
		v.buffer[i] = v.black
	}
}

func (v *VirtualDisplay) SetPixel(x, y int16, c color.RGBA) {
	if x >= 0 && x < v.width && y >= 0 && y < v.height {
		v.buffer[y*v.width+x] = c
	}
}

func (v *VirtualDisplay) Size() (int16, int16) {
	return v.width, v.height
}

func (v *VirtualDisplay) Display() error {
	return nil
}

// NEW: Unified font management methods
func (v *VirtualDisplay) GetFont(fontType font.FontType) font.FontPrinter {
	if v == nil || v.fonts == nil {
		return nil
	}
	return v.fonts.GetFont(fontType)
}

func (v *VirtualDisplay) DrawText(fontType font.FontType, x, y int16, text string) {
	if v == nil || v.fonts == nil {
		return
	}
	font := v.fonts.GetFont(fontType)
	if font != nil {
		font.Print(x, y, text)
	}
}

func (v *VirtualDisplay) CalcTextWidth(fontType font.FontType, text string) int16 {
	if v == nil || v.fonts == nil {
		return 0
	}
	font := v.fonts.GetFont(fontType)
	if font != nil {
		return font.CalcWidth(text)
	}
	return 0
}

func (v *VirtualDisplay) SavePNG(filename string) error {
	img := image.NewRGBA(image.Rect(0, 0, int(v.width), int(v.height)))
	for y := 0; y < int(v.height); y++ {
		for x := 0; x < int(v.width); x++ {
			idx := y*int(v.width) + x
			img.Set(x, y, v.buffer[idx])
		}
	}

	// Resize the image to 5x its original size
	resizedImg := resize.Resize(
		uint(v.width*2),
		uint(v.height*2),
		img,
		resize.NearestNeighbor,
	)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, resizedImg)
}

func (v *VirtualDisplay) DrawXLargeText(x, y int16, text string) {
	v.DrawText(font.FreemonoBold18, x, y, text)
}

func (v *VirtualDisplay) CalcXLargeTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreemonoBold18, text)
}

func (v *VirtualDisplay) DrawLargeText(x, y int16, text string) {
	v.DrawText(font.FreemonoRegular12, x, y, text)
}

func (v *VirtualDisplay) DrawLargeBoldText(x, y int16, text string) {
	v.DrawText(font.FreemonoBold12, x, y, text)
}

func (v *VirtualDisplay) DrawLargeSansText(x, y int16, text string) {
	v.DrawText(font.FreesansRegular12, x, y, text)
}

func (v *VirtualDisplay) CalcLargeTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreemonoRegular12, text)
}

func (v *VirtualDisplay) CalcLargeBoldTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreemonoBold12, text)
}

func (v *VirtualDisplay) CalcLargeSansTextWidth(text string) int16 {
	return v.CalcTextWidth(font.FreesansRegular12, text)
}

func (v *VirtualDisplay) DrawSmallText(x, y int16, text string) {
	v.DrawText(font.ProggySZ8, x, y, text)
}

func (v *VirtualDisplay) CalcSmallTextWidth(text string) int16 {
	return v.CalcTextWidth(font.ProggySZ8, text)
}

// wrapText splits a string into lines at word boundaries to fit a maximum width.
func wrapText(text string, font font.FontPrinter, maxWidth int16) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := words[0]

	for i := 1; i < len(words); i++ {
		word := words[i]
		testLine := currentLine + " " + word

		testWidth := font.CalcWidth(testLine)
		if testWidth > maxWidth {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}
	lines = append(lines, currentLine)
	return lines
}

func (v *VirtualDisplay) DrawLongText(x, y int16, text string) {
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

func (v *VirtualDisplay) DrawPlot(data []int16, title string) {
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

func (v *VirtualDisplay) DrawTwoSideBar(
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

func (v *VirtualDisplay) DrawSparkline(
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

func (v *VirtualDisplay) colorFor(on bool) color.RGBA {
	if on {
		return v.white
	}
	return v.black
}

func (v *VirtualDisplay) FillRect(x, y, w, h int16, on bool) {
	tinydraw.FilledRectangle(v, x, y, w, h, v.colorFor(on))
}

func (v *VirtualDisplay) FillTriangle(x0, y0, x1, y1, x2, y2 int16, on bool) {
	tinydraw.FilledTriangle(v, x0, y0, x1, y1, x2, y2, v.colorFor(on))
}

func (v *VirtualDisplay) DrawSquareBar(x, y int16, value uint8) {
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
