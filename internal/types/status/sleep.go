package status

func SleepStatus(co2 uint16, tempC, dewPointC, absoluteHumidityGM3 float32) string {
	switch {
	case co2 > 1200:
		return "VENT"
	case tempC >= 28:
		return "HOT"
	case dewPointC > 20 || absoluteHumidityGM3 > 17:
		return "WET"
	case co2 > 1000:
		return "VENT?"
	case tempC >= 26.5:
		return "HOT?"
	case dewPointC > 18 || absoluteHumidityGM3 > 15:
		return "WET?"
	default:
		return "OK"
	}
}

func SleepAdvice(status string) (string, string) {
	switch status {
	case "VENT":
		return "OPEN WINDOW", "NOW"
	case "VENT?":
		return "OPEN SMALL GAP", "WATCH CO2"
	case "HOT":
		return "AC COOL 25C", "ECO OFF"
	case "HOT?":
		return "WAIT COOLING", "ECO OFF IF SLOW"
	case "WET":
		return "CLOSE + DRY", "AC/DEHUMID"
	case "WET?":
		return "CLOSE WINDOW", "COOL TO 26C"
	default:
		return "KEEP SETTINGS", "SLEEP"
	}
}
