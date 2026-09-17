package status

import "testing"

func TestSleepStatus(t *testing.T) {
	tests := []struct {
		name                string
		co2                 uint16
		temp, dew, humidity float32
		want                string
	}{
		{"ok", 900, 25.5, 17.9, 14.9, "OK"}, {"ventilation borderline", 1001, 25.5, 17, 14, "VENT?"},
		{"hot before wet borderline", 900, 28, 19, 16, "HOT"}, {"moisture borderline", 900, 25.5, 18.1, 16, "WET?"},
		{"wet by dew point", 900, 25.5, 20.1, 16, "WET"}, {"wet by absolute humidity", 900, 25.5, 18, 17.1, "WET"},
		{"hot before wet", 900, 28.1, 20.1, 17.1, "HOT"}, {"vent priority", 1201, 29, 21, 18, "VENT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SleepStatus(tt.co2, tt.temp, tt.dew, tt.humidity); got != tt.want {
				t.Fatalf("SleepStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSleepAdvice(t *testing.T) {
	tests := []struct{ status, first, second string }{
		{"VENT", "OPEN WINDOW", "NOW"}, {"VENT?", "OPEN SMALL GAP", "WATCH CO2"}, {"HOT", "AC COOL 25C", "ECO OFF"},
		{"HOT?", "WAIT COOLING", "ECO OFF IF SLOW"}, {"WET", "CLOSE + DRY", "AC/DEHUMID"}, {"WET?", "CLOSE WINDOW", "COOL TO 26C"}, {"OK", "KEEP SETTINGS", "SLEEP"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			first, second := SleepAdvice(tt.status)
			if first != tt.first || second != tt.second {
				t.Fatalf("SleepAdvice(%q) = %q, %q", tt.status, first, second)
			}
		})
	}
}
