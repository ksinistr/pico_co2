package display

import (
	"pico_co2/internal/types"
	"testing"
)

func TestGlanceSectionThresholds(t *testing.T) {
	cells := glanceCells(&types.Readings{Raw: types.RawReadings{Temperature: 27, Humidity: 61}})
	for _, tt := range []struct {
		name         string
		cell         glanceCell
		lower, upper float32
	}{
		{"co2", cells[0], 1000, 1900},
		{"temperature", cells[1], 28, 32},
		{"humidity", cells[2], 70, 80},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, geometry := range []struct {
				name         string
				width, count int16
			}{
				{"wide", 4, 5},
				{"narrow", 2, 7},
			} {
				t.Run(geometry.name, func(t *testing.T) {
					for _, reading := range []struct {
						name     string
						value    float32
						sections int16
					}{
						{"below lower", tt.lower - 0.01, 0},
						{"at lower", tt.lower, 1},
						{"middle", (tt.lower + tt.upper) / 2, 1 + (geometry.count-1)/2},
						{"below upper", tt.upper - 0.01, geometry.count - 1},
						{"at upper", tt.upper, geometry.count},
						{"above upper", tt.upper + 1, geometry.count},
					} {
						t.Run(reading.name, func(t *testing.T) {
							device := NewVirtualDisplay(128, 32)
							cell := tt.cell
							cell.value = reading.value
							glanceGaugeSections(device, 91, 0, 38, 7, geometry.width, 3, cell)
							pixels := 0
							for _, pixel := range device.buffer {
								if pixel.R == 255 {
									pixels++
								}
							}
							if want := int(reading.sections * geometry.width * 7); pixels != want {
								t.Fatalf("value %v: lit pixels = %d, want %d", reading.value, pixels, want)
							}
						})
					}
				})
			}
		})
	}
}
