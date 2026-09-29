package widget

import (
	"image/color"
	"pico_co2/pkg/font"
	"testing"

	"tinygo.org/x/drivers"
)

type testDisplay struct {
	width  int16
	height int16
	pixels map[[2]int16]color.RGBA
}

func (d *testDisplay) SetPixel(x, y int16, pixelColor color.RGBA) {
	if d.pixels != nil {
		d.pixels[[2]int16{x, y}] = pixelColor
	}
}

func (d *testDisplay) Size() (int16, int16) { return d.width, d.height }
func (d *testDisplay) Display() error       { return nil }

var _ drivers.Displayer = (*testDisplay)(nil)

func TestTwoSideBarDoesNotAddEmptySideSpacing(t *testing.T) {
	display := &testDisplay{width: 128, height: 32}
	face := font.New(display, font.ProggySZ8)
	for _, tt := range []struct {
		name        string
		left, right int16
		want        int16
	}{
		{"empty left", 0, 2, 39},
		{"empty right", 2, 0, 39},
		{"both empty", 0, 0, 14},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := TwoSideBar(display, face, 0, 0, 0, "H", tt.left, tt.right); got != tt.want {
				t.Fatalf("TwoSideBar() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSquareBar(t *testing.T) {
	for _, tt := range []struct {
		name                               string
		x, width, barHeight, barWidth, gap int16
		value, lower, upper                float32
		wantPixels                         int
	}{
		{
			name:       "uses local x coordinate",
			x:          91,
			width:      38,
			barHeight:  2,
			barWidth:   2,
			gap:        2,
			value:      5,
			lower:      0,
			upper:      10,
			wantPixels: 20,
		},
		{name: "apparent temperature at 27C and 61 percent humidity", width: 38, barHeight: 2, barWidth: 4, gap: 3, value: 30.159, lower: 28, upper: 32, wantPixels: 24},
		{name: "below upper bound stays partial", width: 38, barHeight: 2, barWidth: 4, gap: 3, value: 31.99, lower: 28, upper: 32, wantPixels: 32},
		{name: "at upper bound fills all sections", width: 38, barHeight: 2, barWidth: 4, gap: 3, value: 32, lower: 28, upper: 32, wantPixels: 40},
		{name: "narrow sections at 27C and 61 percent humidity", width: 38, barHeight: 2, barWidth: 2, gap: 3, value: 30.159, lower: 28, upper: 32, wantPixels: 16},
		{
			name:      "below lower bound",
			width:     30,
			barHeight: 2,
			barWidth:  2,
			gap:       10,
			value:     21,
			lower:     25,
			upper:     32,
		},
		{
			name:       "at lower bound",
			wantPixels: 4,
			width:      30,
			barHeight:  2,
			barWidth:   2,
			gap:        10,
			value:      25,
			lower:      25,
			upper:      32,
		},
		{
			name:       "keeps clipped last section out",
			width:      38,
			barHeight:  2,
			barWidth:   2,
			gap:        2,
			value:      10,
			lower:      0,
			upper:      10,
			wantPixels: 36,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			display := &testDisplay{
				width:  128,
				height: 32,
				pixels: make(map[[2]int16]color.RGBA),
			}

			SquareBar(display, tt.x, 0, tt.width, tt.barHeight, tt.barWidth, tt.gap, tt.value, tt.lower, 0, tt.upper)

			if got := countPixels(display, whiteColor()); got != tt.wantPixels {
				t.Fatalf("SquareBar() drew %d pixels, want %d", got, tt.wantPixels)
			}
		})
	}
}

func TestSquareBarIgnoresInvalidInput(t *testing.T) {
	for _, tt := range []struct {
		name                            string
		width, height, barWidth, barGap int16
		lower, upper                    float32
	}{
		{name: "zero width", width: 0, height: 2, barWidth: 2, barGap: 2, lower: 0, upper: 1},
		{name: "zero height", width: 10, height: 0, barWidth: 2, barGap: 2, lower: 0, upper: 1},
		{name: "zero bar width", width: 10, height: 2, barWidth: 0, barGap: 2, lower: 0, upper: 1},
		{name: "negative gap", width: 10, height: 2, barWidth: 2, barGap: -1, lower: 0, upper: 1},
		{name: "equal range", width: 10, height: 2, barWidth: 2, barGap: 2, lower: 1, upper: 1},
		{name: "reversed range", width: 10, height: 2, barWidth: 2, barGap: 2, lower: 2, upper: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			display := &testDisplay{
				width:  128,
				height: 32,
				pixels: make(map[[2]int16]color.RGBA),
			}

			SquareBar(display, 0, 0, tt.width, tt.height, tt.barWidth, tt.barGap, 1, tt.lower, 0, tt.upper)

			if got := countPixels(display, whiteColor()); got != 0 {
				t.Fatalf("SquareBar() drew %d pixels for invalid input", got)
			}
		})
	}

	SquareBar(nil, 0, 0, 10, 2, 2, 2, 1, 0, 0, 1)
}

func countPixels(display *testDisplay, want color.RGBA) int {
	count := 0
	for _, got := range display.pixels {
		if got == want {
			count++
		}
	}

	return count
}
