package display

import (
	"fmt"
	"math"
	"pico_co2/internal/types"
	"pico_co2/internal/types/status"
	"pico_co2/pkg/widget"
)

func FormatClock(clock types.Time) string {
	if clock.LastRead.IsZero() {
		return "--:--"
	}

	return FormatTime(clock.Hour, clock.Minute)
}

func readingText(reading types.Reading, unit string) string {
	switch reading.State {
	case types.ReadingWarmingUp:
		return "WAIT"
	case types.ReadingFailed:
		return "ERR"
	case types.ReadingUnavailable:
		return "--"
	case types.ReadingReady:
	default:
		return "--"
	}

	return fmt.Sprintf("%.0f%s", math.Round(reading.Value), unit)
}

func readingMessage(label string, reading types.Reading) string {
	switch reading.State {
	case types.ReadingReady:
		return ""
	case types.ReadingWarmingUp:
		return label + " warming up"
	case types.ReadingFailed:
		if reading.Err != nil {
			return fmt.Sprintf("%s: %v", label, reading.Err)
		}

		return label + " failed"
	default:
		return "No " + label + " reading"
	}
}

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

func climateReady(readings types.Readings) bool {
	return readings.Temperature.State == types.ReadingReady && readings.Humidity.State == types.ReadingReady
}
