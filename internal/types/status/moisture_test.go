package status

import (
	"math"
	"testing"
)

func TestDewPointC(t *testing.T) {
	tests := []struct {
		name  string
		tempC float32
		rh    float32
		want  float32
	}{
		{"comfortable bedroom", 24, 50, 12.9},
		{"humid bedroom", 24, 90, 22.3},
		{"low rh clamps", 24, 0, -35.6},
		{"high rh clamps", 24, 120, 24.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DewPointC(tt.tempC, tt.rh)
			if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) {
				t.Fatalf("DewPointC(%v, %v) = %v", tt.tempC, tt.rh, got)
			}
			if math.Abs(float64(got-tt.want)) > 0.2 {
				t.Fatalf("DewPointC(%v, %v) = %.2f, want %.2f", tt.tempC, tt.rh, got, tt.want)
			}
		})
	}
}

func TestAbsoluteHumidityGM3(t *testing.T) {
	tests := []struct {
		name  string
		tempC float32
		rh    float32
		want  float32
	}{
		{"comfortable bedroom", 24, 50, 10.9},
		{"humid bedroom", 24, 90, 19.6},
		{"high rh clamps", 24, 120, 21.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AbsoluteHumidityGM3(tt.tempC, tt.rh)
			if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) {
				t.Fatalf("AbsoluteHumidityGM3(%v, %v) = %v", tt.tempC, tt.rh, got)
			}
			if math.Abs(float64(got-tt.want)) > 0.2 {
				t.Fatalf("AbsoluteHumidityGM3(%v, %v) = %.2f, want %.2f", tt.tempC, tt.rh, got, tt.want)
			}
		})
	}
}
