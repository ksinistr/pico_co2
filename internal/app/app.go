package app

import (
	"fmt"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/types"
	"time"

	"tinygo.org/x/drivers"
)

type Schedule struct {
	TimeReadInterval       time.Duration
	StartupReadInterval    time.Duration
	StartupWindow          time.Duration
	SensorReadInterval     time.Duration
	HistoryInterval        time.Duration
	SimultaneousPressDelay time.Duration
}

func DefaultSchedule() Schedule {
	return Schedule{
		TimeReadInterval:       time.Second,
		StartupReadInterval:    time.Second,
		StartupWindow:          time.Minute,
		SensorReadInterval:     time.Minute,
		HistoryInterval:        time.Minute,
		SimultaneousPressDelay: clockedit.DefaultSimultaneousWindow,
	}
}

type Sensors interface {
	Read(now time.Time) types.Readings
}

type Clock interface {
	ReadTime() (time.Time, error)
}

type Deps struct {
	Display      drivers.Displayer
	Screens      []display.Screen
	Clock        Clock
	SetTime      func(time.Time) error
	Sensors      Sensors
	LeftPressed  func() bool
	RightPressed func() bool
	Schedule     Schedule
}

type App struct {
	deps          Deps
	readings      types.Readings
	history       *types.History
	clock         types.Time
	nextTimeRead  time.Time
	clockErr      error
	screen        int
	editor        clockedit.EditorState
	lastLeftTime  time.Time
	lastRightTime time.Time
}

func New(deps Deps, queueCapacity int) *App {
	if len(deps.Screens) == 0 {
		panic("app: at least one screen is required")
	}

	deps.Schedule = scheduleWithDefaults(deps.Schedule)
	deps.Display = display.NewFrameDisplay(deps.Display)

	return &App{
		deps: deps,
		history: types.NewHistory(
			queueCapacity,
			deps.Schedule.HistoryInterval,
		),
	}
}

func scheduleWithDefaults(schedule Schedule) Schedule {
	defaults := DefaultSchedule()
	if schedule.TimeReadInterval == 0 {
		schedule.TimeReadInterval = defaults.TimeReadInterval
	}

	if schedule.StartupReadInterval == 0 {
		schedule.StartupReadInterval = defaults.StartupReadInterval
	}

	if schedule.StartupWindow == 0 {
		schedule.StartupWindow = defaults.StartupWindow
	}

	if schedule.SensorReadInterval == 0 {
		schedule.SensorReadInterval = defaults.SensorReadInterval
	}

	if schedule.HistoryInterval == 0 {
		schedule.HistoryInterval = defaults.HistoryInterval
	}

	if schedule.SimultaneousPressDelay == 0 {
		schedule.SimultaneousPressDelay = defaults.SimultaneousPressDelay
	}

	return schedule
}

func (a *App) Tick(now time.Time) {
	if a.deps.Sensors != nil {
		a.readings = a.deps.Sensors.Read(now)
		a.history.Add(a.readings)
	}

	a.readTime(now)
	a.handleInput(now)
	a.render()
}

func (a *App) readTime(now time.Time) {
	if a.deps.Clock == nil || now.Before(a.nextTimeRead) {
		return
	}

	a.nextTimeRead = now.Add(a.deps.Schedule.TimeReadInterval)

	current, err := a.deps.Clock.ReadTime()
	if err != nil {
		a.clockErr = fmt.Errorf("clock read: %w", err)
		return
	}

	a.clockErr = nil
	a.clock = types.Time{
		Hour:     current.Hour(),
		Minute:   current.Minute(),
		LastRead: now,
	}
}

func (a *App) handleInput(now time.Time) {
	left := pressed(a.deps.LeftPressed)

	right := pressed(a.deps.RightPressed)

	if left {
		a.lastLeftTime = now
	}

	if right {
		a.lastRightTime = now
	}

	result := clockedit.ProcessInput(
		a.editor,
		clockedit.ButtonEvent{
			LeftPressed:  left,
			RightPressed: right,
			LeftTime:     a.lastLeftTime,
			RightTime:    a.lastRightTime,
		},
		a.allowsClockEdit(),
		a.deps.Schedule.SimultaneousPressDelay,
	)

	if result.AttemptEntry {
		a.editor = clockedit.Enter(a.clock.Hour, a.clock.Minute)
		a.resetButtonTimes()

		return
	}

	if result.Exited {
		if result.ShouldSave {
			a.save(result.State, now)
		} else {
			a.editor = clockedit.EditorState{}
		}

		a.resetButtonTimes()

		return
	}

	a.editor = result.State

	if result.NavLeft {
		a.screen = (a.screen - 1 + len(a.deps.Screens)) % len(a.deps.Screens)
		a.lastLeftTime = time.Time{}
	}

	if result.NavRight {
		a.screen = (a.screen + 1) % len(a.deps.Screens)
		a.lastRightTime = time.Time{}
	}
}

func (a *App) save(editor clockedit.EditorState, now time.Time) {
	current, err := a.deps.Clock.ReadTime()
	if err != nil {
		a.editor = clockedit.HandleSaveAttempt(editor, err)
		a.clockErr = fmt.Errorf("clock read: %w", err)

		return
	}

	updated := time.Date(current.Year(), current.Month(), current.Day(),
		editor.Hour, editor.Minute, 0, 0, current.Location())
	err = a.deps.SetTime(updated)

	a.editor = clockedit.HandleSaveAttempt(editor, err)
	if err != nil {
		a.clockErr = fmt.Errorf("clock save: %w", err)
		return
	}

	a.clock.Hour = updated.Hour()
	a.clock.Minute = updated.Minute()
	a.clock.LastRead = now
	a.clockErr = nil
	a.nextTimeRead = now.Add(a.deps.Schedule.TimeReadInterval)
}

func (a *App) render() {
	view := types.ViewState{
		Readings: a.readings,
		History:  a.history,
		CO2Trend: a.history.CO2Trend(),
		Time:     a.clock,
		ClockEdit: types.ClockEdit{
			Active: a.editor.Active,
			Field:  a.editor.Field,
			Hour:   a.editor.Hour,
			Minute: a.editor.Minute,
		},
	}
	if a.clockErr != nil {
		view.Time = types.Time{}

		if a.editor.Active {
			display.RenderError(a.deps.Display, a.clockErr.Error())
			return
		}
	}

	current := a.currentScreen()

	current.Render(a.deps.Display, &view)
}

func (a *App) currentScreen() display.Screen {
	if a.screen >= len(a.deps.Screens) {
		a.screen = 0
	}

	return a.deps.Screens[a.screen]
}

func (a *App) resetButtonTimes() {
	a.lastLeftTime = time.Time{}
	a.lastRightTime = time.Time{}
}

func pressed(read func() bool) bool { return read != nil && read() }

func (a *App) allowsClockEdit() bool {
	return a.currentScreen().AllowsClockEdit && a.deps.Clock != nil && a.deps.SetTime != nil &&
		!a.clock.LastRead.IsZero() &&
		a.clockErr == nil
}
