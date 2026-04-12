// Package rtc defines a narrow interface for real-time clock operations
// and provides pure helpers for building save values. Keeping this logic
// separate from the app package allows host-side testing without TinyGo
// hardware imports.
package rtc

import (
	"fmt"
	"time"
)

// RTC abstracts the real-time clock operations the app needs at runtime.
// Construction and initialization remain in the app's New function; this
// interface covers only the methods used after the device is set up.
type RTC interface {
	// ReadTime returns the current date and time from the RTC hardware.
	ReadTime() (time.Time, error)

	// SetTime writes the given time to the RTC hardware.
	SetTime(time.Time) error
}

// BuildSaveTime constructs the time.Time value to write back to the RTC when
// the user saves from the clock editor.
//
// Strategy: keep the current RTC date (year, month, day, location), replace
// hour and minute from editor state, and reset seconds to zero so the clock
// starts the new minute cleanly.
func BuildSaveTime(current time.Time, hour, minute int) time.Time {
	return time.Date(
		current.Year(),
		current.Month(),
		current.Day(),
		hour,
		minute,
		0, // seconds explicitly reset to zero
		0, // nanoseconds
		current.Location(),
	)
}

// SaveTime writes a new time to the RTC using the editor's pending hour and
// minute values. It builds the save time from the current RTC reading, calls
// SetTime, and returns the new time on success. On failure it wraps the error.
func SaveTime(rtc RTC, hour, minute int, currentTime time.Time) (time.Time, error) {
	newTime := BuildSaveTime(currentTime, hour, minute)
	if err := rtc.SetTime(newTime); err != nil {
		return time.Time{}, fmt.Errorf("rtc set time: %w", err)
	}
	return newTime, nil
}
