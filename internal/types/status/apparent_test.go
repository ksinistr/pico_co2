package status

import (
	"math"
	"testing"
)

func TestApparentTempC(t *testing.T) {
	tests := []struct {
		name  string
		tempC float32
		rh    float32
		want  float32
	}{
		{"cool bedroom", 24, 55, 25.4},
		{"dry at 26", 26, 40, 26.4},
		{"air conditioned", 26, 62, 28.9},
		{"warm", 28, 65, 32.1},
		{"humid at same 27", 27, 96, 34.3},
		{"high rh clamps", 27, 120, 34.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApparentTempC(tt.tempC, tt.rh)
			if math.Abs(float64(got-tt.want)) > 0.2 {
				t.Fatalf("ApparentTempC(%v, %v) = %.2f, want %.2f",
					tt.tempC, tt.rh, got, tt.want)
			}
		})
	}
}

func TestToSleepThermal(t *testing.T) {
	tests := []struct {
		name      string
		apparentC float32
		want      SleepThermal
		wantText  string
	}{
		{"below warm", 27.9, GoodSleepThermal, "GOOD"},
		{"at warm", 28, WarmSleepThermal, "WARM"},
		{"below hot", 31.9, WarmSleepThermal, "WARM"},
		{"at hot", 32, HotSleepThermal, "HOT"},
		{"far above hot", 40, HotSleepThermal, "HOT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSleepThermal(tt.apparentC)
			if got != tt.want {
				t.Fatalf("ToSleepThermal(%v) = %v, want %v", tt.apparentC, got, tt.want)
			}
			if got.String() != tt.wantText {
				t.Fatalf("ToSleepThermal(%v).String() = %q, want %q",
					tt.apparentC, got.String(), tt.wantText)
			}
		})
	}
}
