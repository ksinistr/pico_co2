package app

import (
	"fmt"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/rtc"
	"pico_co2/internal/types"
	"slices"
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

type Deps struct {
	Display      drivers.Displayer
	Screens      []display.Screen
	RTC          rtc.RTC
	Sensors      []Registration
	LeftPressed  func() bool
	RightPressed func() bool
	Schedule     Schedule
}

type App struct {
	deps          Deps
	readings      *types.Readings
	screen        int
	editor        clockedit.EditorState
	lastLeftTime  time.Time
	lastRightTime time.Time
	collector     *Collector
	snapshot      types.Measurements
}

func New(deps Deps, queueCapacity int) *App {
	if len(deps.Screens) == 0 {
		panic("app: at least one screen is required")
	}

	deps.Schedule = scheduleWithDefaults(deps.Schedule)

	output := display.NewFrameDisplay(deps.Display)
	deps.Display = output

	return &App{
		collector: NewCollector(deps.Sensors, queueCapacity, deps.Schedule.HistoryInterval),
		deps:      deps,
		readings: types.InitReadings(
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
	if a.readings == nil || len(a.deps.Screens) == 0 {
		return
	}

	if !a.handleInput(now) {
		a.updateReadings(now)
	}

	a.render()
}

func (a *App) handleInput(now time.Time) bool {
	left := pressed(a.deps.LeftPressed)

	right := pressed(a.deps.RightPressed)
	if left {
		a.lastLeftTime = now
	}

	if right {
		a.lastRightTime = now
	}

	current := a.currentScreen()
	result := clockedit.ProcessInput(a.editor, clockedit.ButtonEvent{
		LeftPressed: left, RightPressed: right, LeftTime: a.lastLeftTime, RightTime: a.lastRightTime,
	}, current.AllowsClockEdit && a.deps.RTC != nil, a.deps.Schedule.SimultaneousPressDelay)

	if result.AttemptEntry {
		a.editor = clockedit.Enter(a.readings.Time.Hour, a.readings.Time.Minute)
		a.resetButtonTimes()

		return true
	}

	if result.Exited {
		if result.ShouldSave {
			a.save(a.deps.RTC, result.State, now)
		} else {
			a.editor = clockedit.EditorState{}
		}

		a.resetButtonTimes()

		return true
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

	return false
}

func (a *App) save(clock rtc.RTC, editor clockedit.EditorState, now time.Time) {
	current, err := clock.ReadTime()
	if err != nil {
		a.editor = clockedit.HandleSaveAttempt(editor, err)
		a.readings.Error = fmt.Sprintf("RTC read: %v", err)

		return
	}

	updated, err := rtc.SaveTime(clock, editor.Hour, editor.Minute, current)

	a.editor = clockedit.HandleSaveAttempt(editor, err)
	if err != nil {
		a.readings.Error = fmt.Sprintf("RTC save: %v", err)
		return
	}

	a.readings.Time.Hour = updated.Hour()
	a.readings.Time.Minute = updated.Minute()
	a.readings.Time.LastRead = now
	a.readings.Error = ""
	a.collector.Invalidate(types.MeasurementTypeTime)
}

func (a *App) updateReadings(now time.Time) {
	values, err := a.collector.Read(now)
	if err != nil {
		a.readings.Error = err.Error()
		return
	}

	a.readings.Error = ""
	if slices.Equal(a.snapshot, values) {
		return
	}

	a.snapshot = values
	a.readings.Measurements = values.Grouped()
	a.readings.SensorHistory = a.collector.History
	a.readings.Project()
}

func (a *App) render() {
	a.readings.ClockEdit = types.ClockEdit{
		Active: a.editor.Active,
		Field:  a.editor.Field,
		Hour:   a.editor.Hour,
		Minute: a.editor.Minute,
	}

	current := a.currentScreen()
	for _, kind := range current.Required {
		if _, ok := a.readings.Measurements[kind].First(); !ok && a.readings.Error == "" {
			display.RenderError(a.deps.Display, fmt.Sprintf("No %s measurement", kind))
			return
		}
	}

	switch {
	case a.readings.Error != "":
		display.RenderError(a.deps.Display, a.readings.Error)
	default:
		a.currentScreen().Render(a.deps.Display, a.readings)
	}
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
