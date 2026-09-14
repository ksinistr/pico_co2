package status

import "math"

func DewPointC(tempC, rh float32) float32 {
	clampedRH := clampRH(rh)
	const (
		a = 17.62
		b = 243.12
	)
	gamma := math.Log(float64(clampedRH)/100.0) + (a*float64(tempC))/(b+float64(tempC))
	return float32((b * gamma) / (a - gamma))
}

func AbsoluteHumidityGM3(tempC, rh float32) float32 {
	clampedRH := clampRH(rh)
	svpHPA := 6.112 * math.Exp((17.62*float64(tempC))/(243.12+float64(tempC)))
	vaporPressure := svpHPA * float64(clampedRH) / 100.0
	return float32(216.7 * vaporPressure / (273.15 + float64(tempC)))
}

func clampRH(rh float32) float32 {
	switch {
	case rh < 1:
		return 1
	case rh > 100:
		return 100
	default:
		return rh
	}
}
