package widget

import (
	"image/color"
	"testing"

	"pico_co2/pkg/font"
	"tinygo.org/x/drivers"
)

type testDisplay struct {
	width  int16
	height int16
}

func (d *testDisplay) SetPixel(int16, int16, color.RGBA) {}
func (d *testDisplay) Size() (int16, int16)              { return d.width, d.height }
func (d *testDisplay) Display() error                    { return nil }

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
