package widget

import "testing"

func TestScaleX(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  int16
	}{
		{"below the scale clamps", 20, 0}, {"scale start", 24, 0}, {"comfort boundary", 28, 42}, {"hot threshold", 32, 85}, {"above the scale clamps", 40, 128},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScaleX(128, tt.value, 24, 36); got != tt.want {
				t.Fatalf("ScaleX() = %d, want %d", got, tt.want)
			}
		})
	}
}
