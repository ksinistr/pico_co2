package app

import (
	"errors"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/sensors"
	"pico_co2/internal/types"
	"testing"
	"time"

	"tinygo.org/x/drivers"
)

type fakeRTC struct {
	now     time.Time
	readErr error
	setErr  error
	saved   time.Time
}

func (r *fakeRTC) ReadTime() (time.Time, error) { return r.now, r.readErr }

func (r *fakeRTC) SetTime(value time.Time) error {
	if r.setErr != nil {
		return r.setErr
	}
	r.saved = value
	return nil
}

func TestTickReadsSensorsOnSchedule(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	for _, tt := range []struct {
		name  string
		steps []time.Duration
		want  int
	}{
		{"first read", []time.Duration{0}, 1},
		{"startup interval", []time.Duration{0, 500 * time.Millisecond, time.Second}, 2},
		{"regular interval", []time.Duration{0, time.Minute}, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reads := 0
			deps := testDependencies(&fakeRTC{now: now}, func() ([]types.Measurement, error) {
				reads++
				return []types.Measurement{{Type: types.MeasurementTypeCO2, Value: 800}}, nil
			})
			state := New(deps, 16)
			for _, offset := range tt.steps {
				state.Tick(now.Add(offset))
			}
			if reads != tt.want {
				t.Fatalf("reads = %d, want %d", reads, tt.want)
			}
		})
	}
}

func TestTickInputAndSave(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	for _, tt := range []struct {
		name  string
		setup func(*App, *fakeRTC, *bool, *bool)
		check func(*testing.T, *App, *fakeRTC)
	}{
		{
			name:  "navigation",
			setup: func(_ *App, _ *fakeRTC, _ *bool, right *bool) { *right = true },
			check: func(t *testing.T, state *App, _ *fakeRTC) {
				if state.screen != 1 {
					t.Fatalf("screen = %d, want 1", state.screen)
				}
			},
		},
		{
			name:  "enter and cancel editor",
			setup: func(_ *App, _ *fakeRTC, left, right *bool) { *left, *right = true, true },
			check: func(t *testing.T, state *App, _ *fakeRTC) {
				if !state.editor.Active {
					t.Fatal("editor is inactive")
				}
			},
		},
		{
			name: "save editor time",
			setup: func(state *App, _ *fakeRTC, _ *bool, right *bool) {
				state.editor.Active = true
				state.editor.Field = clockedit.FieldSave
				state.editor.Hour = 8
				state.editor.Minute = 15
				*right = true
			},
			check: func(t *testing.T, state *App, clock *fakeRTC) {
				if state.editor.Active {
					t.Fatal("editor is active after save")
				}
				if clock.saved.Hour() != 8 || clock.saved.Minute() != 15 {
					t.Fatalf("saved = %v", clock.saved)
				}
			},
		},
		{
			name: "save error reopens editor",
			setup: func(state *App, clock *fakeRTC, _ *bool, right *bool) {
				state.editor.Active = true
				state.editor.Field = clockedit.FieldSave
				state.editor.Hour = 8
				state.editor.Minute = 15
				clock.setErr = errors.New("write failed")
				*right = true
			},
			check: func(t *testing.T, state *App, _ *fakeRTC) {
				if !state.editor.Active || state.readings.Error == "" {
					t.Fatalf("editor active = %t, error = %q", state.editor.Active, state.readings.Error)
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			left, right := false, false
			clock := &fakeRTC{now: now}
			deps := testDependencies(clock, func() ([]types.Measurement, error) { return []types.Measurement{{Type: types.MeasurementTypeCO2}}, nil })
			deps.LeftPressed = func() bool { value := left; left = false; return value }
			deps.RightPressed = func() bool { value := right; right = false; return value }
			state := New(deps, 16)
			state.readings.Time.Hour, state.readings.Time.Minute = 14, 23
			tt.setup(state, clock, &left, &right)
			state.Tick(now)
			tt.check(t, state, clock)
		})
	}
}

func TestTickRecoversFromSensorError(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	reads := 0
	deps := testDependencies(&fakeRTC{now: now}, func() ([]types.Measurement, error) {
		reads++
		if reads == 1 {
			return nil, errors.New("sensor unavailable")
		}
		return []types.Measurement{{Type: types.MeasurementTypeCO2, Value: 800}}, nil
	})
	state := New(deps, 16)
	state.Tick(now)
	if state.readings.Error == "" {
		t.Fatal("missing sensor error")
	}
	state.Tick(now.Add(time.Second))
	if state.readings.Error != "" {
		t.Fatalf("error = %q", state.readings.Error)
	}
}

func TestTickShowsErrorWhenSensorsFail(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	rendered := false
	deps := testDependencies(&fakeRTC{now: now}, func() ([]types.Measurement, error) {
		return nil, errors.New("sensor unavailable")
	})
	deps.Screens[0].Render = func(drivers.Displayer, *types.Readings) { rendered = true }
	monitor := New(deps, 16)
	monitor.editor = clockedit.Enter(14, 23)

	monitor.Tick(now)

	if rendered || monitor.readings.Error == "" {
		t.Fatal("sensor error did not override editor")
	}
}

func testDependencies(clock *fakeRTC, read readerFunc) Deps {
	return Deps{
		Display: display.NewVirtualDisplay(128, 32),
		Screens: []display.Screen{
			{ID: "time", AllowsClockEdit: true, Render: func(drivers.Displayer, *types.Readings) {}},
			{ID: "second", Render: func(drivers.Displayer, *types.Readings) {}},
		},
		RTC: clock,
		Sensors: []Registration{
			{ID: "test", Reader: read, Capabilities: []types.MeasurementType{types.MeasurementTypeCO2}, Interval: time.Minute, StartupInterval: time.Second, StartupWindow: time.Minute},
			{ID: "rtc", Reader: sensors.Clock{Source: clock}, Capabilities: []types.MeasurementType{types.MeasurementTypeTime}, Interval: time.Second},
		},
		Schedule: Schedule{
			TimeReadInterval:       time.Second,
			StartupReadInterval:    time.Second,
			StartupWindow:          time.Minute,
			SensorReadInterval:     time.Minute,
			HistoryInterval:        time.Minute,
			SimultaneousPressDelay: 400 * time.Millisecond,
		},
	}
}

type readerFunc func() ([]types.Measurement, error)

func (f readerFunc) Read() ([]types.Measurement, error) { return f() }

type countedDisplay struct {
	*display.VirtualDisplay
	sends int
}

func (d *countedDisplay) Display() error { d.sends++; return nil }

func TestTickFrameChanges(t *testing.T) {
	now := time.Date(2026, 10, 21, 14, 23, 0, 0, time.UTC)
	clock := &fakeRTC{now: now}
	target := &countedDisplay{VirtualDisplay: display.NewVirtualDisplay(128, 32)}
	failing := false
	deps := testDependencies(clock, func() ([]types.Measurement, error) {
		if failing {
			return nil, errors.New("disconnected")
		}
		return []types.Measurement{{Type: types.MeasurementTypeCO2, Value: 800}}, nil
	})
	deps.Display = target
	deps.Screens[0].Render = display.RenderTime
	deps.Sensors[0].Interval = time.Second
	state := New(deps, 16)
	for _, tt := range []struct {
		name   string
		offset time.Duration
		fail   bool
		minute int
		sends  int
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
			clock.now = time.Date(2026, 10, 21, 14, tt.minute, 0, 0, time.UTC)
			state.Tick(now.Add(tt.offset))
			if target.sends != tt.sends {
				t.Fatalf("display sends=%d want=%d", target.sends, tt.sends)
			}
		})
	}
}
