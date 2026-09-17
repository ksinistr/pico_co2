package widget

import "testing"

func TestScaleX(t *testing.T) {
	tests := []struct {
		name  string
		width int16
		value float32
		min   float32
		max   float32
		want  int16
	}{
		{"below the scale clamps", 128, 20, 24, 36, 0},
		{"scale start", 128, 24, 24, 36, 0},
		{"comfort boundary", 128, 28, 24, 36, 42},
		{"hot threshold", 128, 32, 24, 36, 85},
		{"above the scale clamps", 128, 40, 24, 36, 128},
		{"zero width", 0, 28, 24, 36, 0},
		{"empty range", 128, 28, 24, 24, 0},
		{"reversed range", 128, 28, 36, 24, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScaleX(tt.width, tt.value, tt.min, tt.max); got != tt.want {
				t.Fatalf("ScaleX() = %d, want %d", got, tt.want)
			}
		})
	}
}
