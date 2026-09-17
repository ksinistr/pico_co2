package font

import (
	"testing"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/notosans"
)

func TestNotosansWidthUsesGlyphMetrics(t *testing.T) {
	for _, tt := range []struct {
		name string
		text string
	}{
		{"digits", "123"},
		{"different glyph widths", "1W"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			face := New(nil, Notosans)
			_, want := tinyfont.LineWidth(&notosans.Notosans12pt, tt.text)
			if got := face.Width(tt.text); got != int16(want) {
				t.Fatalf("Width(%q) = %d, want %d", tt.text, got, want)
			}
		})
	}
}
