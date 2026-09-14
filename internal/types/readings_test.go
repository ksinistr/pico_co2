package types

import (
	"math"
	"testing"
	"time"

	"pico_co2/internal/types/status"
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
			r := InitReadings(128)

			// Add readings one by one to simulate real usage
			for i, co2 := range tt.readings {
				// Manually advance time first to bypass the granularity check
				if i > 0 {
					r.History.AddedAt = r.History.AddedAt.Add(-time.Duration(i) * time.Minute)
				}
				r.AddReadings(co2, 22.0, 50.0)
			}

			// Check the trend
			if r.Calculated.CO2Trend != tt.expectedTrend {
				t.Errorf("Expected trend %v, got %v", tt.expectedTrend, r.Calculated.CO2Trend)
			}

			// Print debug info for failed tests
			if r.Calculated.CO2Trend != tt.expectedTrend {
				t.Logf("Debug info - Previous avg: %d, Current avg: %d, Diff: %d",
					r.Calculated.CO25MinAvgPrev, r.Calculated.CO25MinAvgCurr,
					int32(r.Calculated.CO25MinAvgCurr)-int32(r.Calculated.CO25MinAvgPrev))
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

func TestAddReadingsCalculatesMoisture(t *testing.T) {
	r := InitReadings(128)

	r.AddReadings(900, 24.0, 90.0)

	if math.Abs(float64(r.Calculated.DewPointC-22.3)) > 0.2 {
		t.Fatalf("DewPointC = %.2f, want around 22.3", r.Calculated.DewPointC)
	}
	if math.Abs(float64(r.Calculated.AbsoluteHumidityGM3-19.6)) > 0.2 {
		t.Fatalf("AbsoluteHumidityGM3 = %.2f, want around 19.6", r.Calculated.AbsoluteHumidityGM3)
	}
}

func TestAddReadingsStoresDewPointHistory(t *testing.T) {
	r := InitReadings(128)
	r.History.AddedAt = time.Now().Add(-2 * time.Minute)

	r.AddReadings(900, 24.0, 90.0)

	data := r.History.DewPoint.Contiguous()
	if len(data) != 1 {
		t.Fatalf("dew point history len = %d, want 1", len(data))
	}
	if data[0] != 22 {
		t.Fatalf("dew point history value = %d, want 22", data[0])
	}
}
