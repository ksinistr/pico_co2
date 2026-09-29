package status

import "math"

// HeatIndexVal is the full NWS algorithm: the Steadman average below 80F, the
// Rothfusz regression with its low and high humidity adjustments above it.
// https://en.wikipedia.org/wiki/Heat_index#Formula
func HeatIndexVal(tempC, rh float32) float32 {
	t := float64(tempC)*9/5 + 32
	r := float64(rh)

	simple := 0.5 * (t + 61 + (t-68)*1.2 + r*0.094)
	if (simple+t)/2 < 80 {
		return fahrenheitToC(simple)
	}

	hi := -42.379 + 2.04901523*t + 10.14333127*r -
		0.22475541*t*r - 0.00683783*t*t - 0.05481717*r*r +
		0.00122874*t*t*r + 0.00085282*t*r*r - 0.00000199*t*t*r*r

	switch {
	case r < 13 && t >= 80 && t <= 112:
		hi -= (13 - r) / 4 * math.Sqrt((17-math.Abs(t-95))/17)
	case r > 85 && t >= 80 && t <= 87:
		hi += (r - 85) / 10 * (87 - t) / 5
	}

	return fahrenheitToC(hi)
}

func fahrenheitToC(f float64) float32 {
	return float32((f - 32) * 5 / 9)
}

type HeatIndex uint8

const (
	NoHeat HeatIndex = iota
	Caution
	ExtremeCaution
	Danger
	ExtremeDanger
	UnknownHeatIndex
)

func GetHeatIndex(tempC, rh float32) HeatIndex {
	heatIndex := HeatIndexVal(tempC, rh)
	return calculateHeatIndex(heatIndex)
}

func calculateHeatIndex(heatIndex float32) HeatIndex {
	switch {
	case heatIndex < 27:
		return NoHeat
	case heatIndex < 32:
		return Caution
	case heatIndex < 41:
		return ExtremeCaution
	case heatIndex < 54:
		return Danger
	default:
		return ExtremeDanger
	}
}

func (h HeatIndex) String() string {
	switch h {
	case NoHeat:
		return "No heat"
	case Caution:
		return "Caution"
	case ExtremeCaution:
		return "Extreme caution"
	case Danger:
		return "Danger"
	case ExtremeDanger:
		return "Extreme danger"
	default:
		return "Unknown Heat Index"
	}
}
