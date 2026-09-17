package layout

import (
	"image/color"
	"reflect"
	"testing"

	"pico_co2/pkg/font"
)

type testDisplay struct {
	width  int16
	height int16
	whiteY []int16
}

func (d *testDisplay) SetPixel(_ int16, y int16, c color.RGBA) {
	if c.R > 0 || c.G > 0 || c.B > 0 {
		d.whiteY = append(d.whiteY, y)
	}
}

func (d *testDisplay) Size() (int16, int16) { return d.width, d.height }
func (d *testDisplay) Display() error       { return nil }

func TestRowPositions(t *testing.T) {
	for _, tt := range []struct {
		name  string
		cells func(*testDisplay) []Cell
		want  func(*testDisplay) []int16
	}{
		{
			name: "three equal faces",
			cells: func(display *testDisplay) []Cell {
				face := font.New(display, font.ProggySZ8)
				return []Cell{{Face: face, Text: "A"}, {Face: face, Text: "BB"}, {Face: face, Text: "C"}}
			},
			want: func(display *testDisplay) []int16 {
				face := font.New(display, font.ProggySZ8)
				gap := (display.width - face.Width("A") - face.Width("BB") - face.Width("C")) / 2
				return []int16{0, face.Width("A") + gap, display.width - face.Width("C")}
			},
		},
		{
			name: "mixed faces and units preserve right alignment",
			cells: func(display *testDisplay) []Cell {
				small := font.New(display, font.ProggySZ8)
				medium := font.New(display, font.FreemonoRegular9)
				large := font.New(display, font.FreemonoRegular12)
				return []Cell{
					{Face: medium, Text: "26", Unit: "c", UnitFace: small},
					{Face: medium, Text: "55", Unit: "%", UnitFace: small},
					{Face: large, Text: "1200", YOffset: -2},
				}
			},
			want: func(display *testDisplay) []int16 {
				small := font.New(display, font.ProggySZ8)
				medium := font.New(display, font.FreemonoRegular9)
				large := font.New(display, font.FreemonoRegular12)
				leftWidth := medium.Width("26") + small.Width("c")
				middleWidth := medium.Width("55") + small.Width("%")
				rightWidth := large.Width("1200")
				gap := (display.width - leftWidth - middleWidth - rightWidth) / 2
				return []int16{0, leftWidth + gap, display.width - rightWidth}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			display := &testDisplay{width: 128, height: 32}
			if got, want := Row(display, 17, tt.cells(display)...), tt.want(display); !reflect.DeepEqual(got, want) {
				t.Fatalf("Row() = %v, want %v", got, want)
			}
		})
	}
}

func TestRowAppliesYOffset(t *testing.T) {
	baselineDisplay := &testDisplay{width: 128, height: 32}
	Row(baselineDisplay, 10, Cell{Face: font.New(baselineDisplay, font.ProggySZ8), Text: "A"})
	baseline := minY(baselineDisplay.whiteY)

	for _, tt := range []struct {
		name   string
		offset int16
	}{
		{"move up", -2},
		{"unchanged", 0},
		{"move down", 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			display := &testDisplay{width: 128, height: 32}
			Row(display, 10, Cell{Face: font.New(display, font.ProggySZ8), Text: "A", YOffset: tt.offset})
			if got, want := minY(display.whiteY), baseline+tt.offset; got != want {
				t.Fatalf("minimum y = %d, want %d", got, want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	display := &testDisplay{width: 128, height: 32}
	face := font.New(display, font.ProggySZ8)
	for _, tt := range []struct {
		name  string
		text  string
		width int16
		want  []string
	}{
		{"words", "sensor timeout while reading", 84, []string{"sensor timeout", "while reading"}},
		{"long word", "abcdefghijkl", 24, []string{"abcd", "efgh", "ijkl"}},
		{"whitespace", "  sensor   timeout ", 84, []string{"sensor timeout"}},
		{"empty", "  ", 84, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wrap(tt.text, face, tt.width); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Wrap(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestLongTextStartsAtOrigin(t *testing.T) {
	for _, tt := range []struct {
		name string
		y    int16
	}{
		{"top edge", 0},
		{"inset", 5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			direct := &testDisplay{width: 128, height: 32}
			font.New(direct, font.ProggySZ8).Print(0, tt.y, "A")

			wrapped := &testDisplay{width: 128, height: 32}
			LongText(wrapped, font.New(wrapped, font.ProggySZ8), 0, tt.y, "A B")

			if got, want := minY(wrapped.whiteY), minY(direct.whiteY); got != want {
				t.Fatalf("minimum y = %d, want %d", got, want)
			}
		})
	}
}

func minY(values []int16) int16 {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}
