// Package rtc separates real-time clock hardware from application logic so
// saving the clock can be tested without a board.
package rtc

import (
	"fmt"
	"time"
)

type RTC interface {
	ReadTime() (time.Time, error)
	SetTime(time.Time) error
}

// BuildSaveTime keeps the RTC date and location while replacing the edited
// hour and minute. Seconds are reset so the new minute starts exactly.
func BuildSaveTime(current time.Time, hour, minute int) time.Time {
	return time.Date(current.Year(), current.Month(), current.Day(),
		hour, minute, 0, 0, current.Location())
}

func SaveTime(rtc RTC, hour, minute int, currentTime time.Time) (time.Time, error) {
	newTime := BuildSaveTime(currentTime, hour, minute)
	if err := rtc.SetTime(newTime); err != nil {
		return time.Time{}, fmt.Errorf("rtc set time: %w", err)
	}
	return newTime, nil
}
