package display

import "testing"

func TestScaleX(t *testing.T) {
	tests := []struct {
		name      string
		apparentC float32
		want      int16
	}{
		{"below the scale clamps", 20, 0},
		{"scale start", 24, 0},
		{"comfort boundary", 28, 42},
		{"hot threshold", 32, 85},
		{"above the scale clamps", 40, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScaleX(128, tt.apparentC)
			if got != tt.want {
				t.Fatalf("ScaleX(128, %v) = %d, want %d", tt.apparentC, got, tt.want)
			}
		})
	}
}
