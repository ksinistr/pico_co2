package rtc

import (
	"errors"
	"testing"
	"time"
)

// mockRTC is a test double for the RTC interface.
type mockRTC struct {
	readTimeFn func() (time.Time, error)
	setTimeFn  func(time.Time) error
}

func (m *mockRTC) ReadTime() (time.Time, error) {
	return m.readTimeFn()
}

func (m *mockRTC) SetTime(t time.Time) error {
	return m.setTimeFn(t)
}

// ---------------------------------------------------------------------------
// BuildSaveTime tests
// ---------------------------------------------------------------------------

func TestBuildSaveTime(t *testing.T) {
	tests := []struct {
		name     string
		current  time.Time
		hour     int
		minute   int
		wantTime string // formatted as "2006-01-02 15:04:05"
	}{
		{
			name:     "replace hour and minute",
			current:  time.Date(2025, 12, 21, 14, 35, 42, 0, time.UTC),
			hour:     9,
			minute:   5,
			wantTime: "2025-12-21 09:05:00",
		},
		{
			name:     "seconds reset to zero",
			current:  time.Date(2026, 1, 15, 23, 59, 58, 0, time.UTC),
			hour:     23,
			minute:   59,
			wantTime: "2026-01-15 23:59:00",
		},
		{
			name:     "midnight",
			current:  time.Date(2026, 3, 1, 15, 30, 10, 0, time.UTC),
			hour:     0,
			minute:   0,
			wantTime: "2026-03-01 00:00:00",
		},
		{
			name:     "preserves date and location",
			current:  time.Date(2026, 6, 30, 10, 0, 30, 500, time.UTC),
			hour:     18,
			minute:   45,
			wantTime: "2026-06-30 18:45:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSaveTime(tt.current, tt.hour, tt.minute)
			gotStr := got.Format("2006-01-02 15:04:05")
			if gotStr != tt.wantTime {
				t.Errorf("BuildSaveTime() = %v, want %v", gotStr, tt.wantTime)
			}
			// Verify date is preserved.
			if got.Year() != tt.current.Year() || got.Month() != tt.current.Month() || got.Day() != tt.current.Day() {
				t.Errorf("date changed: got %d-%02d-%02d, want %d-%02d-%02d",
					got.Year(), got.Month(), got.Day(),
					tt.current.Year(), tt.current.Month(), tt.current.Day())
			}
			// Verify seconds are zero.
			if got.Second() != 0 {
				t.Errorf("seconds not reset: got %d", got.Second())
			}
			// Verify nanoseconds are zero.
			if got.Nanosecond() != 0 {
				t.Errorf("nanoseconds not reset: got %d", got.Nanosecond())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SaveTime success tests
// ---------------------------------------------------------------------------

func TestSaveTimeSuccess(t *testing.T) {
	tests := []struct {
		name        string
		currentTime time.Time
		hour        int
		minute      int
		wantHour    int
		wantMinute  int
	}{
		{
			name:        "basic save",
			currentTime: time.Date(2025, 12, 21, 14, 35, 42, 0, time.UTC),
			hour:        9,
			minute:      5,
			wantHour:    9,
			wantMinute:  5,
		},
		{
			name:        "save midnight",
			currentTime: time.Date(2026, 1, 1, 23, 59, 59, 0, time.UTC),
			hour:        0,
			minute:      0,
			wantHour:    0,
			wantMinute:  0,
		},
		{
			name:        "save max values",
			currentTime: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			hour:        23,
			minute:      59,
			wantHour:    23,
			wantMinute:  59,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedTime time.Time
			mock := &mockRTC{
				readTimeFn: func() (time.Time, error) {
					return tt.currentTime, nil
				},
				setTimeFn: func(t time.Time) error {
					receivedTime = t
					return nil
				},
			}

			result, err := SaveTime(mock, tt.hour, tt.minute, tt.currentTime)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Hour() != tt.wantHour {
				t.Errorf("result hour = %d, want %d", result.Hour(), tt.wantHour)
			}
			if result.Minute() != tt.wantMinute {
				t.Errorf("result minute = %d, want %d", result.Minute(), tt.wantMinute)
			}
			if result.Second() != 0 {
				t.Errorf("seconds not reset: got %d", result.Second())
			}
			// Verify the value that was actually passed to SetTime.
			if receivedTime.Hour() != tt.wantHour {
				t.Errorf("SetTime received hour = %d, want %d", receivedTime.Hour(), tt.wantHour)
			}
			if receivedTime.Minute() != tt.wantMinute {
				t.Errorf("SetTime received minute = %d, want %d", receivedTime.Minute(), tt.wantMinute)
			}
			if receivedTime.Second() != 0 {
				t.Errorf("SetTime received seconds = %d, want 0", receivedTime.Second())
			}
			// Date must be preserved from the current RTC time.
			if receivedTime.Year() != tt.currentTime.Year() ||
				receivedTime.Month() != tt.currentTime.Month() ||
				receivedTime.Day() != tt.currentTime.Day() {
				t.Errorf("date not preserved: got %d-%02d-%02d, want %d-%02d-%02d",
					receivedTime.Year(), receivedTime.Month(), receivedTime.Day(),
					tt.currentTime.Year(), tt.currentTime.Month(), tt.currentTime.Day())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SaveTime failure tests
// ---------------------------------------------------------------------------

func TestSaveTimeFailure(t *testing.T) {
	tests := []struct {
		name    string
		setErr  error
		wantErr string
	}{
		{
			name:    "i2c write error",
			setErr:  errors.New("i2c write: device not responding"),
			wantErr: "rtc set time: i2c write: device not responding",
		},
		{
			name:    "generic hardware error",
			setErr:  errors.New("hardware fault"),
			wantErr: "rtc set time: hardware fault",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRTC{
				setTimeFn: func(t time.Time) error {
					return tt.setErr
				},
			}

			current := time.Date(2025, 12, 21, 14, 35, 0, 0, time.UTC)
			_, err := SaveTime(mock, 9, 5, current)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// mockRTC compliance tests
// ---------------------------------------------------------------------------

func TestMockRTCReadTime(t *testing.T) {
	expected := time.Date(2026, 4, 12, 10, 30, 0, 0, time.UTC)
	mock := &mockRTC{
		readTimeFn: func() (time.Time, error) {
			return expected, nil
		},
	}
	got, err := mock.ReadTime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(expected) {
		t.Errorf("ReadTime() = %v, want %v", got, expected)
	}
}

func TestMockRTCReadTimeError(t *testing.T) {
	wantErr := errors.New("i2c read failed")
	mock := &mockRTC{
		readTimeFn: func() (time.Time, error) {
			return time.Time{}, wantErr
		},
	}
	_, err := mock.ReadTime()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != wantErr.Error() {
		t.Errorf("error = %q, want %q", err.Error(), wantErr.Error())
	}
}
