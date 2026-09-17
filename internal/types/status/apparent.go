package status

// Steadman apparent temperature without the wind term. Unlike the heat index in
// heatindex.go it stays defined below 27C, which is the range an air
// conditioned bedroom lives in. See docs/thermal-index.md.
func ApparentTempC(tempC, rh float32) float32 {
	return tempC + 0.33*VaporPressureHPA(tempC, rh) - 4
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

var SleepThermalStrings = [...]string{
	"GOOD",
	"WARM",
	"HOT",
}

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
	if s > HotSleepThermal {
		return SleepThermalStrings[HotSleepThermal]
	}
	return SleepThermalStrings[s]
}
