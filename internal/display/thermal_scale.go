package display

import "pico_co2/internal/types/status"

const (
	scaleMinApparentC float32 = 24
	scaleMaxApparentC float32 = 36
)

// ScaleX maps an apparent temperature to a column, clamped to the scale ends.
func ScaleX(width int16, apparentC float32) int16 {
	switch {
	case apparentC < scaleMinApparentC:
		apparentC = scaleMinApparentC
	case apparentC > scaleMaxApparentC:
		apparentC = scaleMaxApparentC
	}

	span := scaleMaxApparentC - scaleMinApparentC
	return int16(float32(width) * (apparentC - scaleMinApparentC) / span)
}

// DrawThermalScale keeps the comfortable range as a bare ruler line and grows a
// solid block only past the boundary: an empty scale means nothing to do, and
// the block length is how far beyond the limit the room has gone. The notch
// inside the block marks the point where sleep is documented to break down.
func DrawThermalScale(renderer Renderer, y, h int16, apparentC float32) {
	if renderer == nil {
		return
	}

	width, _ := renderer.Size()
	origin := ScaleX(width, status.WarmApparentC)

	renderer.FillRect(0, y+h-1, origin, 1, true)
	renderer.FillRect(origin, y-2, 1, h+2, true)

	value := ScaleX(width, apparentC)
	if value <= origin {
		return
	}

	renderer.FillRect(origin, y, value-origin, h, true)

	hot := ScaleX(width, status.HotApparentC)
	renderer.FillRect(hot-1, y, 3, h, false)
	renderer.FillRect(hot, y, 1, h, true)
}
