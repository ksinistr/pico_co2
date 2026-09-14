package display

import "testing"

func TestSleepStatus(t *testing.T) {
	tests := []struct {
		name                string
		co2                 uint16
		tempC               float32
		dewPointC           float32
		absoluteHumidityGM3 float32
		want                string
	}{
		{"ok", 900, 25.5, 17.9, 14.9, "OK"},
		{"ventilation borderline", 1001, 25.5, 17.0, 14.0, "VENT?"},
		{"hot before wet borderline", 900, 28.0, 19.0, 16.0, "HOT"},
		{"moisture borderline", 900, 25.5, 18.1, 16.0, "WET?"},
		{"wet by dew point", 900, 25.5, 20.1, 16.0, "WET"},
		{"wet by absolute humidity", 900, 25.5, 18.0, 17.1, "WET"},
		{"hot before wet", 900, 28.1, 20.1, 17.1, "HOT"},
		{"vent priority", 1201, 29.0, 21.0, 18.0, "VENT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SleepStatus(tt.co2, tt.tempC, tt.dewPointC, tt.absoluteHumidityGM3)
			if got != tt.want {
				t.Fatalf("SleepStatus(%d, %.1f, %.1f, %.1f) = %q, want %q",
					tt.co2, tt.tempC, tt.dewPointC, tt.absoluteHumidityGM3, got, tt.want)
			}
		})
	}
}

func TestSleepAdvice(t *testing.T) {
	tests := []struct {
		status string
		line1  string
		line2  string
	}{
		{"VENT", "OPEN WINDOW", "NOW"},
		{"VENT?", "OPEN SMALL GAP", "WATCH CO2"},
		{"HOT", "AC COOL 25C", "ECO OFF"},
		{"HOT?", "WAIT COOLING", "ECO OFF IF SLOW"},
		{"WET", "CLOSE + DRY", "AC/DEHUMID"},
		{"WET?", "CLOSE WINDOW", "COOL TO 26C"},
		{"OK", "KEEP SETTINGS", "SLEEP"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			line1, line2 := SleepAdvice(tt.status)
			if line1 != tt.line1 || line2 != tt.line2 {
				t.Fatalf("SleepAdvice(%q) = %q, %q, want %q, %q",
					tt.status, line1, line2, tt.line1, tt.line2)
			}
		})
	}
}
