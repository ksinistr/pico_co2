package rtc

import "time"

type RTC interface {
	ReadTime() (time.Time, error)
	SetTime(time.Time) error
}
