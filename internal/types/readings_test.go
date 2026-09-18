package types

import (
	"pico_co2/internal/types/status"
	"testing"
	"time"
)

func TestCO2TrendCalculation(t *testing.T) {
	tests := []struct {
		name          string
		readings      []uint16
		expectedTrend status.CO2Trend
	}{
		{
			name:          "Insufficient data",
			readings:      []uint16{400, 410, 420},
			expectedTrend: status.UnknownCO2Trend,
		},
		{
			name:          "Stable readings",
			readings:      []uint16{400, 405, 395, 400, 410, 408, 402, 398, 405, 403},
			expectedTrend: status.StableCO2,
		},
		{
			name:          "Rising trend - exactly +50 ppm",
			readings:      []uint16{400, 405, 410, 415, 420, 450, 455, 460, 465, 470},
			expectedTrend: status.StableCO2, // exactly 50ppm should be stable
		},
		{
			name:          "Rising trend - more than +50 ppm",
			readings:      []uint16{400, 405, 410, 415, 420, 480, 485, 490, 495, 500},
			expectedTrend: status.RisingCO2,
		},
		{
			name:          "Falling trend - more than -50 ppm",
			readings:      []uint16{500, 495, 490, 485, 480, 420, 415, 410, 405, 400},
			expectedTrend: status.FallingCO2,
		},
		{
			name:          "Falling trend - exactly -50 ppm",
			readings:      []uint16{500, 495, 490, 485, 480, 470, 465, 460, 455, 450},
			expectedTrend: status.StableCO2, // exactly -50ppm should be stable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InitReadings(128, time.Minute)
			start := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
			for i, co2 := range tt.readings {
				r.AddReadingsAt(start.Add(time.Duration(i)*time.Minute), co2, 22.0, 50.0)
			}

			if r.Calculated.CO2Trend != tt.expectedTrend {
				t.Errorf("Expected trend %v, got %v", tt.expectedTrend, r.Calculated.CO2Trend)
			}
		})
	}
}

func TestHistoryRecordsAtIntervalBoundary(t *testing.T) {
	for _, tt := range []struct {
		name    string
		offsets []time.Duration
		want    int
	}{
		{"before boundary", []time.Duration{0, 59 * time.Second}, 1},
		{"at boundary", []time.Duration{0, time.Minute}, 2},
		{"multiple intervals", []time.Duration{0, time.Minute, 2 * time.Minute}, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			readings := InitReadings(8, time.Minute)
			start := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
			for _, offset := range tt.offsets {
				readings.AddReadingsAt(start.Add(offset), 800, 24, 50)
			}
			if got := readings.History.CO2.Len(); got != tt.want {
				t.Fatalf("history length = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCO2TrendString(t *testing.T) {
	tests := []struct {
		trend    status.CO2Trend
		expected string
	}{
		{status.StableCO2, "Stable"},
		{status.RisingCO2, "Rising"},
		{status.FallingCO2, "Falling"},
		{status.UnknownCO2Trend, "Unknown"},
		{status.CO2Trend(99), "Unknown"}, // invalid value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.trend.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.trend.String())
			}
		})
	}
}
