package app

import (
	"errors"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/types"
	"testing"
	"time"

	"tinygo.org/x/drivers"
)

type fakeClock struct {
	now     time.Time
	readErr error
	setErr  error
	saved   time.Time
	reads   int
}

func (r *fakeClock) ReadTime() (time.Time, error) { r.reads++; return r.now, r.readErr }
func (r *fakeClock) SetTime(value time.Time) error {
	if r.setErr != nil {
		return r.setErr
	}
	r.saved = value
	r.now = value
	return nil
}

type clockFunc func() (time.Time, error)

func (f clockFunc) ReadTime() (time.Time, error) { return f() }

type readerFunc func(time.Time) types.Readings

func (f readerFunc) Read(now time.Time) types.Readings { return f(now) }

func readyReadings(now time.Time) types.Readings {
	return types.Readings{
		CO2:         types.Reading{Value: 800, UpdatedAt: now, State: types.ReadingReady},
		Temperature: types.Reading{Value: 24, UpdatedAt: now, State: types.ReadingReady},
		Humidity:    types.Reading{Value: 55, UpdatedAt: now, State: types.ReadingReady},
	}
}

func testDependencies() Deps {
	return Deps{
		Display: display.NewVirtualDisplay(128, 32),
		Screens: []display.Screen{
			{ID: "time", AllowsClockEdit: true, Render: func(drivers.Displayer, *types.ViewState) {}},
			{ID: "second", Render: func(drivers.Displayer, *types.ViewState) {}},
		},
		Sensors: readerFunc(readyReadings),
	}
}

func TestTickClockAndHistory(t *testing.T) {
	now := time.Date(2026, 10, 28, 14, 23, 0, 0, time.UTC)
	for _, tt := range []struct {
		name        string
		clock       bool
		clockFails  bool
		sensorFails bool
		cached      bool
	}{
		{name: "normal", clock: true},
		{name: "without clock"},
		{name: "clock fails independently", clock: true, clockFails: true},
		{name: "sensor fails independently", clock: true, sensorFails: true},
		{name: "cached snapshot", clock: true, cached: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			deps := testDependencies()
			clock := &fakeClock{now: now}
			if tt.clock {
				deps.Clock = clockFunc(clock.ReadTime)
			}
			if tt.clockFails {
				clock.readErr = errors.New("clock disconnected")
			}
			var captured types.ViewState
			deps.Screens[0].Render = func(_ drivers.Displayer, view *types.ViewState) { captured = *view }
			deps.Sensors = readerFunc(func(at time.Time) types.Readings {
				if tt.cached {
					at = now
				}
				readings := readyReadings(at)
				if tt.sensorFails {
					readings.CO2 = types.Reading{State: types.ReadingFailed, Err: errors.New("sensor disconnected")}
				}
				return readings
			})
			state := New(deps, 16)
			for _, offset := range []time.Duration{0, 500 * time.Millisecond, time.Second, time.Minute} {
				state.Tick(now.Add(offset))
			}
			wantCO2 := 2
			if tt.cached {
				wantCO2 = 1
			}
			if tt.sensorFails {
				wantCO2 = 0
			}
			if state.history.CO2.Values.Len() != wantCO2 {
				t.Fatalf("CO2 history length = %d", state.history.CO2.Values.Len())
			}
			if captured.Readings.Temperature.State != types.ReadingReady {
				t.Fatal("lost available temperature")
			}
			if (tt.clock && !tt.clockFails) != !captured.Time.LastRead.IsZero() {
				t.Fatal("wrong clock availability")
			}
			if tt.clock && clock.reads != 3 {
				t.Fatalf("clock reads = %d, want 3", clock.reads)
			}
		})
	}
}

func TestTickInputAndSave(t *testing.T) {
	now := time.Date(2026, 10, 28, 14, 23, 0, 0, time.UTC)
	for _, tt := range []struct {
		name  string
		setup func(*App, *fakeClock, *bool, *bool)
		check func(*testing.T, *App, *fakeClock)
	}{
		{
			name:  "navigation",
			setup: func(_ *App, _ *fakeClock, _ *bool, right *bool) { *right = true },
			check: func(t *testing.T, state *App, _ *fakeClock) {
				if state.screen != 1 {
					t.Fatal("did not navigate")
				}
			},
		},
		{
			name:  "enter editor",
			setup: func(_ *App, _ *fakeClock, left, right *bool) { *left, *right = true, true },
			check: func(t *testing.T, state *App, _ *fakeClock) {
				if !state.editor.Active || state.editor.Hour != 14 || state.editor.Minute != 23 {
					t.Fatal("wrong editor initial time")
				}
			},
		},
		{
			name: "cancel editor",
			setup: func(state *App, _ *fakeClock, left, right *bool) {
				state.editor = clockedit.Enter(8, 15)
				*left, *right = true, true
			},
			check: func(t *testing.T, state *App, clock *fakeClock) {
				if state.editor.Active || !clock.saved.IsZero() {
					t.Fatal("cancel saved or kept editor")
				}
			},
		},
		{
			name: "save time",
			setup: func(state *App, _ *fakeClock, _ *bool, right *bool) {
				state.editor = clockedit.Enter(8, 15)
				state.editor.Field = clockedit.FieldSave
				*right = true
			},
			check: func(t *testing.T, state *App, clock *fakeClock) {
				if state.editor.Active || clock.saved.Hour() != 8 || clock.saved.Minute() != 15 || state.clock.Hour != 8 {
					t.Fatal("did not save time")
				}
				if clock.saved.Day() != now.Day() || clock.saved.Second() != 0 {
					t.Fatal("changed date or did not reset seconds")
				}
			},
		},
		{
			name: "save error",
			setup: func(state *App, clock *fakeClock, _ *bool, right *bool) {
				state.editor = clockedit.Enter(8, 15)
				state.editor.Field = clockedit.FieldSave
				clock.setErr = errors.New("write failed")
				*right = true
			},
			check: func(t *testing.T, state *App, _ *fakeClock) {
				if !state.editor.Active || state.clockErr == nil {
					t.Fatal("missing save error")
				}
			},
		},
		{
			name: "clock read error",
			setup: func(_ *App, clock *fakeClock, left, right *bool) {
				clock.readErr = errors.New("read failed")
				*left, *right = true, true
			},
			check: func(t *testing.T, state *App, _ *fakeClock) {
				if state.editor.Active || state.clockErr == nil {
					t.Fatal("entered editor without clock")
				}
			},
		},
		{
			name:  "without clock",
			setup: func(state *App, _ *fakeClock, left, right *bool) { state.deps.Clock = nil; *left, *right = true, true },
			check: func(t *testing.T, state *App, _ *fakeClock) {
				if state.editor.Active {
					t.Fatal("entered editor without clock")
				}
			},
		},
		{
			name: "read-only clock",
			setup: func(state *App, clock *fakeClock, left, right *bool) {
				state.deps.Clock = clockFunc(clock.ReadTime)
				state.deps.SetTime = nil
				*left, *right = true, true
			},
			check: func(t *testing.T, state *App, clock *fakeClock) {
				if state.editor.Active || state.clock.LastRead.IsZero() || !clock.saved.IsZero() {
					t.Fatal("read-only clock was edited or did not provide time")
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			left, right := false, false
			clock := &fakeClock{now: now}
			deps := testDependencies()
			deps.Clock = clock
			deps.SetTime = clock.SetTime
			deps.LeftPressed = func() bool { value := left; left = false; return value }
			deps.RightPressed = func() bool { value := right; right = false; return value }
			state := New(deps, 16)
			tt.setup(state, clock, &left, &right)
			state.Tick(now)
			tt.check(t, state, clock)
		})
	}
}

type countedDisplay struct {
	*display.VirtualDisplay
	sends int
}

func (d *countedDisplay) Display() error { d.sends++; return nil }

func TestTickFrameChanges(t *testing.T) {
	now := time.Date(2026, 10, 28, 14, 23, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	target := &countedDisplay{VirtualDisplay: display.NewVirtualDisplay(128, 32)}
	failing := false
	deps := testDependencies()
	deps.Clock, deps.Display = clock, target
	deps.Screens[0].Render = display.RenderTime
	deps.Sensors = readerFunc(func(at time.Time) types.Readings {
		readings := readyReadings(at)
		if failing {
			readings.CO2.State = types.ReadingFailed
		}
		return readings
	})
	state := New(deps, 16)
	for _, tt := range []struct {
		name          string
		offset        time.Duration
		fail          bool
		minute, sends int
	}{
		{"initial", 0, false, 23, 1},
		{"new sample same pixels", time.Second, false, 23, 1},
		{"clock changed", 2 * time.Second, false, 24, 2},
		{"error appears", 3 * time.Second, true, 24, 3},
		{"same error stays", 4 * time.Second, true, 24, 3},
		{"recovery", 5 * time.Second, false, 24, 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			failing = tt.fail
			clock.now = time.Date(2026, 10, 28, 14, tt.minute, 0, 0, time.UTC)
			state.Tick(now.Add(tt.offset))
			if target.sends != tt.sends {
				t.Fatalf("display sends = %d, want %d", target.sends, tt.sends)
			}
		})
	}
}

func TestSavePreservesDateAndLocation(t *testing.T) {
	for _, tt := range []struct {
		name         string
		current      time.Time
		hour, minute int
	}{
		{"midnight", time.Date(2026, 12, 31, 23, 59, 58, 500, time.UTC), 0, 0},
		{"end of day", time.Date(2026, 1, 1, 0, 0, 30, 500, time.UTC), 23, 59},
		{"local time", time.Date(2026, 10, 28, 14, 23, 45, 500, time.FixedZone("UTC+2", 2*60*60)), 8, 15},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clock := &fakeClock{now: tt.current}
			deps := testDependencies()
			deps.Clock = clock
			deps.SetTime = clock.SetTime
			state := New(deps, 8)
			state.save(clockedit.Enter(tt.hour, tt.minute), tt.current)
			want := time.Date(tt.current.Year(), tt.current.Month(), tt.current.Day(),
				tt.hour, tt.minute, 0, 0, tt.current.Location())
			if !clock.saved.Equal(want) || clock.saved.Location() != tt.current.Location() {
				t.Fatalf("saved time = %v, want %v", clock.saved, want)
			}
			if state.clock.Hour != tt.hour || state.clock.Minute != tt.minute || state.clockErr != nil {
				t.Fatal("saved time was not reflected in the application")
			}
		})
	}
}
