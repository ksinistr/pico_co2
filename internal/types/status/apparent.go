package status

import "math"

// Steadman apparent temperature without the wind term. Unlike the heat index in
// heatindex.go it stays defined below 27C, which is the range an air
// conditioned bedroom lives in. See docs/thermal-index.md.
func ApparentTempC(tempC, rh float32) float32 {
	return tempC + 0.33*VaporPressureHPA(tempC, rh) - 4
}

func VaporPressureHPA(tempC, rh float32) float32 {
	if rh < 1 {
		rh = 1
	}

	if rh > 100 {
		rh = 100
	}

	svpHPA := 6.112 * math.Exp((17.62*float64(tempC))/(243.12+float64(tempC)))

	return float32(svpHPA * float64(rh) / 100)
}

// Thresholds are apparent temperatures, converted from the dry-bulb values the
// sleep literature reports at 50-60% RH: 26C/55% -> 28.1, 29C/55% -> 32.3.
const (
	WarmApparentC float32 = 28
	HotApparentC  float32 = 32
)

type SleepThermal uint8

const (
	GoodSleepThermal SleepThermal = iota
	WarmSleepThermal
	HotSleepThermal
)

func ToSleepThermal(apparentC float32) SleepThermal {
	switch {
	case apparentC < WarmApparentC:
		return GoodSleepThermal
	case apparentC < HotApparentC:
		return WarmSleepThermal
	default:
		return HotSleepThermal
	}
}

func (s SleepThermal) String() string {
	switch s {
	case GoodSleepThermal:
		return "GOOD"
	case WarmSleepThermal:
		return "WARM"
	case HotSleepThermal:
		return "HOT"
	default:
		return "HOT"
	}
}
