package display

import (
	"image/color"

	"tinygo.org/x/drivers"
)

type VirtualDisplay struct {
	buffer []color.RGBA
	width  int16
	height int16
}

func NewVirtualDisplay(width, height int16) *VirtualDisplay {
	return &VirtualDisplay{
		buffer: make([]color.RGBA, width*height),
		width:  width,
		height: height,
	}
}

func (v *VirtualDisplay) Clear() {
	for i := range v.buffer {
		v.buffer[i] = color.RGBA{0, 0, 0, 255}
	}
}

func (v *VirtualDisplay) SetPixel(x, y int16, pixelColor color.RGBA) {
	if x >= 0 && x < v.width && y >= 0 && y < v.height {
		v.buffer[y*v.width+x] = pixelColor
	}
}

//nolint:nonamedreturns // The drivers.Displayer interface requires two int16 results.
func (v *VirtualDisplay) Size() (width, height int16) { return v.width, v.height }

func (*VirtualDisplay) Display() error { return nil }

var _ drivers.Displayer = (*VirtualDisplay)(nil)
