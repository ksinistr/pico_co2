package status

import (
	"math"
	"testing"
)

func TestHeatIndexVal(t *testing.T) {
	tests := []struct {
		name  string
		tempC float32
		rh    float32
		want  float32
	}{
		{"cool bedroom", 24, 55, 23.9},
		{"below regression", 26, 61, 26.2},
		{"no step at 27", 27, 61, 28.2},
		{"warm", 28, 61, 29.6},
		{"humid", 30, 70, 35.0},
		{"very humid adjustment", 27, 96, 32.2},
		{"very dry adjustment", 38, 10, 34.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HeatIndexVal(tt.tempC, tt.rh)
			if math.Abs(float64(got-tt.want)) > 0.2 {
				t.Fatalf("HeatIndexVal(%v, %v) = %.2f, want %.2f",
					tt.tempC, tt.rh, got, tt.want)
			}
		})
	}
}
