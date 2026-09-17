package app

import (
	"fmt"
	"time"

	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/rtc"
	"pico_co2/internal/sensors"
	"pico_co2/internal/types"
	"tinygo.org/x/drivers"
)

const TimeScreenID = "time"

type Dependencies struct {
	Display                drivers.Displayer
	Screens                []display.Screen
	RTC                    rtc.RTC
	Read                   sensors.Reader
	LeftPressed            func() bool
	RightPressed           func() bool
	TimeReadInterval       time.Duration
	StartupReadInterval    time.Duration
	SensorReadInterval     time.Duration
	SimultaneousPressDelay time.Duration
}

type State struct {
	Readings      *types.Readings
	Screen        int
	Editor        clockedit.EditorState
	lastLeftTime  time.Time
	lastRightTime time.Time
}

func NewState(queueCapacity int) *State {
	return &State{Readings: types.InitReadings(queueCapacity)}
}

func (s *State) Step(deps Dependencies, now time.Time) {
	if s.Readings == nil || len(deps.Screens) == 0 {
		return
	}
	if !s.handleInput(deps, now) {
		s.updateReadings(deps, now)
	}
	s.render(deps)
}

func (s *State) handleInput(deps Dependencies, now time.Time) bool {
	left := pressed(deps.LeftPressed)
	right := pressed(deps.RightPressed)
	if left {
		s.lastLeftTime = now
	}
	if right {
		s.lastRightTime = now
	}

	result := clockedit.ProcessInput(s.Editor, clockedit.ButtonEvent{
		LeftPressed: left, RightPressed: right, LeftTime: s.lastLeftTime, RightTime: s.lastRightTime,
	}, s.currentScreen(deps).ID == TimeScreenID, deps.SimultaneousPressDelay)

	if result.AttemptEntry {
		s.Editor = clockedit.Enter(s.Readings.Time.Hour, s.Readings.Time.Minute)
		s.resetButtonTimes()
		s.Readings.IsDrawen = false
		return true
	}
	if result.Exited {
		if result.ShouldSave {
			s.save(deps.RTC, result.State, now)
		} else {
			s.Editor = clockedit.EditorState{}
		}
		s.resetButtonTimes()
		s.Readings.IsDrawen = false
		return true
	}

	s.Editor = result.State
	if result.NavLeft {
		s.Screen = (s.Screen - 1 + len(deps.Screens)) % len(deps.Screens)
		s.lastLeftTime = time.Time{}
	}
	if result.NavRight {
		s.Screen = (s.Screen + 1) % len(deps.Screens)
		s.lastRightTime = time.Time{}
	}
	if result.RedrawNeeded {
		s.Readings.IsDrawen = false
	}
	return false
}

func (s *State) save(clock rtc.RTC, editor clockedit.EditorState, now time.Time) {
	current, err := clock.ReadTime()
	if err != nil {
		s.Editor = clockedit.HandleSaveAttempt(editor, err)
		s.Readings.Error = fmt.Sprintf("RTC read: %v", err)
		return
	}
	updated, err := rtc.SaveTime(clock, editor.Hour, editor.Minute, current)
	s.Editor = clockedit.HandleSaveAttempt(editor, err)
	if err != nil {
		s.Readings.Error = fmt.Sprintf("RTC save: %v", err)
		return
	}
	s.Readings.Time.Hour = updated.Hour()
	s.Readings.Time.Minute = updated.Minute()
	s.Readings.Time.LastRead = now
	s.Readings.Error = ""
}

func (s *State) updateReadings(deps Dependencies, now time.Time) {
	if s.Readings.Time.LastRead.IsZero() || now.Sub(s.Readings.Time.LastRead) >= deps.TimeReadInterval {
		current, err := deps.RTC.ReadTime()
		if err != nil {
			s.Readings.Error = fmt.Sprintf("DS3231: %v", err)
			s.Readings.IsDrawen = false
		} else {
			s.Readings.Time.LastRead = now
			if s.Readings.Time.Hour != current.Hour() || s.Readings.Time.Minute != current.Minute() {
				s.Readings.Time.Hour = current.Hour()
				s.Readings.Time.Minute = current.Minute()
				s.Readings.IsDrawen = false
			}
		}
	}
	if !s.shouldReadSensors(deps, now) {
		return
	}
	raw, err := deps.Read()
	if err != nil {
		s.Readings.Error = err.Error()
		s.Readings.IsDrawen = false
		return
	}
	s.Readings.AddReadingsAt(now, raw.CO2, raw.Temperature, raw.Humidity)
	s.Readings.IsDrawen = false
}

func (s *State) shouldReadSensors(deps Dependencies, now time.Time) bool {
	if s.Readings.LastUpdateAt.IsZero() {
		return true
	}
	elapsed := now.Sub(s.Readings.LastUpdateAt)
	if now.Sub(s.Readings.FirstReadingAt) < time.Minute {
		return elapsed >= deps.StartupReadInterval
	}
	return elapsed >= deps.SensorReadInterval
}

func (s *State) render(deps Dependencies) {
	if s.Readings.IsDrawen {
		return
	}
	s.Readings.ClockEdit = types.ClockEdit{Active: s.Editor.Active, Field: int(s.Editor.Field), Hour: s.Editor.Hour, Minute: s.Editor.Minute}
	if s.Readings.Error != "" {
		display.RenderError(deps.Display, s.Readings)
	} else {
		s.currentScreen(deps).Render(deps.Display, s.Readings)
	}
	s.Readings.IsDrawen = true
}

func (s *State) currentScreen(deps Dependencies) display.Screen {
	if s.Screen >= len(deps.Screens) {
		s.Screen = 0
	}
	return deps.Screens[s.Screen]
}

func (s *State) resetButtonTimes() {
	s.lastLeftTime = time.Time{}
	s.lastRightTime = time.Time{}
}

func pressed(read func() bool) bool { return read != nil && read() }
