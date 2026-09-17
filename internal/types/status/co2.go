package status

type CO2Trend uint8

const (
	StableCO2 CO2Trend = iota
	RisingCO2
	FallingCO2
	UnknownCO2Trend
)

var CO2TrendStrings = [...]string{
	"Stable",
	"Rising",
	"Falling",
	"Unknown",
}

func (c CO2Trend) String() string {
	if c < StableCO2 || c > UnknownCO2Trend {
		return "Unknown"
	}
	return CO2TrendStrings[c]
}
