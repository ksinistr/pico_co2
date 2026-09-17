package app

import (
	"errors"
	"testing"
	"time"

	"pico_co2/internal/display"
	"pico_co2/internal/sensors"
	"pico_co2/internal/types"
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

func TestStepReadsSensorsOnSchedule(t *testing.T) {
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
			deps := testDependencies(&fakeRTC{now: now}, func() (sensors.Raw, error) {
				reads++
				return sensors.Raw{CO2: 800, Temperature: 24, Humidity: 50}, nil
			})
			state := NewState(16)
			for _, offset := range tt.steps {
				state.Step(deps, now.Add(offset))
			}
			if reads != tt.want {
				t.Fatalf("reads = %d, want %d", reads, tt.want)
			}
		})
	}
}

func TestStepInputAndSave(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	for _, tt := range []struct {
		name  string
		setup func(*State, *fakeRTC, *bool, *bool)
		check func(*testing.T, *State, *fakeRTC)
	}{
		{
			name:  "navigation",
			setup: func(_ *State, _ *fakeRTC, _ *bool, right *bool) { *right = true },
			check: func(t *testing.T, state *State, _ *fakeRTC) {
				if state.Screen != 1 {
					t.Fatalf("screen = %d, want 1", state.Screen)
				}
			},
		},
		{
			name:  "enter and cancel editor",
			setup: func(_ *State, _ *fakeRTC, left, right *bool) { *left, *right = true, true },
			check: func(t *testing.T, state *State, _ *fakeRTC) {
				if !state.Editor.Active {
					t.Fatal("editor is inactive")
				}
			},
		},
		{
			name: "save editor time",
			setup: func(state *State, _ *fakeRTC, _ *bool, right *bool) {
				state.Editor.Active = true
				state.Editor.Field = 2
				state.Editor.Hour = 8
				state.Editor.Minute = 15
				*right = true
			},
			check: func(t *testing.T, state *State, clock *fakeRTC) {
				if state.Editor.Active {
					t.Fatal("editor is active after save")
				}
				if clock.saved.Hour() != 8 || clock.saved.Minute() != 15 {
					t.Fatalf("saved = %v", clock.saved)
				}
			},
		},
		{
			name: "save error reopens editor",
			setup: func(state *State, clock *fakeRTC, _ *bool, right *bool) {
				state.Editor.Active = true
				state.Editor.Field = 2
				state.Editor.Hour = 8
				state.Editor.Minute = 15
				clock.setErr = errors.New("write failed")
				*right = true
			},
			check: func(t *testing.T, state *State, _ *fakeRTC) {
				if !state.Editor.Active || state.Readings.Error == "" {
					t.Fatalf("editor active = %t, error = %q", state.Editor.Active, state.Readings.Error)
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			left, right := false, false
			clock := &fakeRTC{now: now}
			state := NewState(16)
			state.Readings.Time.Hour, state.Readings.Time.Minute = 14, 23
			tt.setup(state, clock, &left, &right)
			deps := testDependencies(clock, func() (sensors.Raw, error) { return sensors.Raw{}, nil })
			deps.LeftPressed = func() bool { value := left; left = false; return value }
			deps.RightPressed = func() bool { value := right; right = false; return value }
			state.Step(deps, now)
			tt.check(t, state, clock)
		})
	}
}

func TestStepRecoversFromSensorError(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	reads := 0
	deps := testDependencies(&fakeRTC{now: now}, func() (sensors.Raw, error) {
		reads++
		if reads == 1 {
			return sensors.Raw{}, errors.New("sensor unavailable")
		}
		return sensors.Raw{CO2: 800, Temperature: 24, Humidity: 50}, nil
	})
	state := NewState(16)
	state.Step(deps, now)
	if state.Readings.Error == "" {
		t.Fatal("missing sensor error")
	}
	state.Step(deps, now.Add(time.Second))
	if state.Readings.Error != "" {
		t.Fatalf("error = %q", state.Readings.Error)
	}
}

func testDependencies(clock *fakeRTC, read sensors.Reader) Dependencies {
	return Dependencies{
		Display: display.NewVirtualDisplay(128, 32),
		Screens: []display.Screen{
			{ID: TimeScreenID, Render: func(drivers.Displayer, *types.Readings) {}},
			{ID: "second", Render: func(drivers.Displayer, *types.Readings) {}},
		},
		RTC:                    clock,
		Read:                   read,
		TimeReadInterval:       time.Second,
		StartupReadInterval:    time.Second,
		SensorReadInterval:     time.Minute,
		SimultaneousPressDelay: 400 * time.Millisecond,
	}
}
