package types

import (
	"math"
	"pico_co2/internal/types/status"
	"testing"
	"time"
)

func TestHistory(t *testing.T) {
	for _, tt := range []struct {
		name     string
		capacity int
	}{
		{"single value", 1},
		{"multiple values", 8},
	} {
		t.Run(tt.name, func(t *testing.T) {
			history := NewHistory(tt.capacity, time.Minute)
			now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			reading := Reading{Value: 24, State: ReadingReady, UpdatedAt: now}
			history.Add(Readings{Temperature: reading, Humidity: reading, CO2: reading})
			for _, series := range []Series{history.Temperature, history.Humidity, history.CO2} {
				if series.Values.Len() != 1 || series.Values.Contiguous()[0] != 24 {
					t.Fatal("missing reading in history")
				}
			}
			if history.CO2Trend() != status.UnknownCO2Trend {
				t.Fatal("trend with insufficient history")
			}
		})
	}
}

func TestHistorySkipsInvalidAndCachedReadings(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name      string
		state     ReadingState
		value     float64
		timestamp time.Time
		want      int
	}{
		{"cached", ReadingReady, 900, now, 1},
		{"older", ReadingReady, 900, now.Add(-time.Minute), 1},
		{"new", ReadingReady, 900, now.Add(time.Minute), 2},
		{"unavailable", ReadingUnavailable, 900, now.Add(time.Minute), 1},
		{"warmup", ReadingWarmingUp, 900, now.Add(time.Minute), 1},
		{"failed", ReadingFailed, 900, now.Add(time.Minute), 1},
		{"zero timestamp", ReadingReady, 900, time.Time{}, 1},
		{"not finite", ReadingReady, math.NaN(), now.Add(time.Minute), 1},
		{"overflow", ReadingReady, 40000, now.Add(time.Minute), 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			history := NewHistory(4, time.Minute)
			history.Add(Readings{CO2: Reading{Value: 800, UpdatedAt: now, State: ReadingReady}})
			history.Add(Readings{CO2: Reading{Value: tt.value, UpdatedAt: tt.timestamp, State: tt.state}, Temperature: Reading{Value: 0, UpdatedAt: now.Add(time.Minute), State: ReadingReady}})
			if history.CO2.Values.Len() != tt.want {
				t.Fatalf("CO2 length = %d", history.CO2.Values.Len())
			}
			if history.Temperature.Values.Len() != 1 || history.Temperature.Values.Contiguous()[0] != 0 {
				t.Fatal("zero temperature was lost or histories were coupled")
			}
		})
	}
}
