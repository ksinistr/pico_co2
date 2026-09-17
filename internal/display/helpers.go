package display

import (
	"pico_co2/internal/types/status"
	"pico_co2/pkg/widget"
)

func trendDirection(trend status.CO2Trend) widget.TrendDirection {
	switch trend {
	case status.RisingCO2:
		return widget.TrendRising
	case status.FallingCO2:
		return widget.TrendFalling
	case status.StableCO2:
		return widget.TrendStable
	default:
		return widget.TrendUnknown
	}
}
