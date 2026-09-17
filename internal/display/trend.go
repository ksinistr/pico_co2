package display

import "pico_co2/internal/types/status"

const trendSize int16 = 4

// DrawTrend marks the CO2 direction at the given centre point: a triangle
// pointing the way it moves, a bar when it holds.
func DrawTrend(renderer Renderer, x, y int16, trend status.CO2Trend) {
	if renderer == nil {
		return
	}

	switch trend {
	case status.RisingCO2:
		renderer.FillTriangle(
			x-trendSize, y+trendSize,
			x+trendSize, y+trendSize,
			x, y-trendSize,
			true,
		)
	case status.FallingCO2:
		renderer.FillTriangle(
			x-trendSize, y-trendSize,
			x+trendSize, y-trendSize,
			x, y+trendSize,
			true,
		)
	case status.StableCO2:
		renderer.FillRect(x-trendSize, y-1, 2*trendSize, 2, true)
	}
}
