package display

import (
	"fmt"
	"image/color"

	"tinygo.org/x/drivers"
)

type FrameDisplay struct {
	target      drivers.Displayer
	frame       *VirtualDisplay
	fingerprint uint64
	sent        bool
}

func NewFrameDisplay(target drivers.Displayer) *FrameDisplay {
	width, height := target.Size()
	return &FrameDisplay{target: target, frame: NewVirtualDisplay(width, height)}
}

//nolint:nonamedreturns // Matches the display dimensions.
func (f *FrameDisplay) Size() (width, height int16)           { return f.frame.Size() }
func (f *FrameDisplay) SetPixel(x, y int16, value color.RGBA) { f.frame.SetPixel(x, y, value) }
func (f *FrameDisplay) Display() error {
	fingerprint := uint64(14695981039346656037)

	for _, pixel := range f.frame.buffer {
		for _, value := range []uint8{pixel.R, pixel.G, pixel.B, pixel.A} {
			fingerprint ^= uint64(value)
			fingerprint *= 1099511628211
		}
	}

	if f.sent && fingerprint == f.fingerprint {
		return nil
	}

	width, height := f.Size()
	for y := range height {
		for x := range width {
			f.target.SetPixel(x, y, f.frame.buffer[y*width+x])
		}
	}

	if err := f.target.Display(); err != nil {
		return fmt.Errorf("display frame: %w", err)
	}

	f.fingerprint = fingerprint
	f.sent = true

	return nil
}
