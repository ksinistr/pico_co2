package status

type CO2Trend uint8

const (
	StableCO2 CO2Trend = iota
	RisingCO2
	FallingCO2
	UnknownCO2Trend
)

func (c CO2Trend) String() string {
	switch c {
	case StableCO2:
		return "Stable"
	case RisingCO2:
		return "Rising"
	case FallingCO2:
		return "Falling"
	case UnknownCO2Trend:
		return "Unknown"
	default:
		return "Unknown"
	}
}
