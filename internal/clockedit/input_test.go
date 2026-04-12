package clockedit

import (
	"errors"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// DetectSimultaneous tests
// ---------------------------------------------------------------------------

func TestDetectSimultaneous(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	window := 400 * time.Millisecond

	tests := []struct {
		name string
		ev   ButtonEvent
		want bool
	}{
		{
			"both pressed same time",
			ButtonEvent{true, true, base, base},
			true,
		},
		{
			"both pressed within window",
			ButtonEvent{true, true, base, base.Add(200 * time.Millisecond)},
			true,
		},
		{
			"both pressed exactly at window edge",
			ButtonEvent{true, true, base, base.Add(400 * time.Millisecond)},
			true,
		},
		{
			"both pressed outside window",
			ButtonEvent{true, true, base, base.Add(401 * time.Millisecond)},
			false,
		},
		{
			"both pressed far apart",
			ButtonEvent{true, true, base, base.Add(2 * time.Second)},
			false,
		},
		{
			"only left pressed",
			ButtonEvent{true, false, base, time.Time{}},
			false,
		},
		{
			"only right pressed",
			ButtonEvent{false, true, time.Time{}, base},
			false,
		},
		{
			"neither pressed",
			ButtonEvent{false, false, time.Time{}, time.Time{}},
			false,
		},
		{
			"both pressed but left time zero",
			ButtonEvent{true, true, time.Time{}, base},
			false,
		},
		{
			"both pressed but right time zero",
			ButtonEvent{true, true, base, time.Time{}},
			false,
		},
		{
			"left pressed before right within window",
			ButtonEvent{true, true, base.Add(-300 * time.Millisecond), base},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectSimultaneous(tt.ev, window)
			if got != tt.want {
				t.Errorf("DetectSimultaneous() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Input handling success paths and redraw behavior
// ---------------------------------------------------------------------------

func TestInputSuccessPaths(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	window := DefaultSimultaneousWindow

	bothEv := ButtonEvent{true, true, base, base}
	leftEv := ButtonEvent{true, false, base, time.Time{}}
	rightEv := ButtonEvent{false, true, time.Time{}, base}
	noneEv := ButtonEvent{false, false, time.Time{}, time.Time{}}

	tests := []struct {
		name         string
		state        EditorState
		ev           ButtonEvent
		onTimeScreen bool
		wantAttempt  bool
		wantNavLeft  bool
		wantNavRight bool
		wantSave     bool
		wantExit     bool
		wantRedraw   bool
		wantActive   bool // expected editor active state after processing
		wantHour     int
		wantMinute   int
		wantField    EditField
	}{
		{
			name:         "enter edit from simultaneous press on time screen",
			state:        EditorState{},
			ev:           bothEv,
			onTimeScreen: true,
			wantAttempt:  true,
			wantRedraw:   true,
		},
		{
			name:        "left navigates previous when not editing",
			state:       EditorState{},
			ev:          leftEv,
			wantNavLeft: true,
			wantRedraw:  true,
		},
		{
			name:         "right navigates next when not editing",
			state:        EditorState{},
			ev:           rightEv,
			wantNavRight: true,
			wantRedraw:   true,
		},
		{
			name:       "editor left cycles field from hour to minute",
			state:      EditorState{Active: true, Field: FieldHour, Hour: 10, Minute: 30},
			ev:         leftEv,
			wantRedraw: true,
			wantActive: true,
			wantHour:   10,
			wantMinute: 30,
			wantField:  FieldMinute,
		},
		{
			name:       "editor left cycles from cancel back to hour",
			state:      EditorState{Active: true, Field: FieldCancel, Hour: 10, Minute: 30},
			ev:         leftEv,
			wantRedraw: true,
			wantActive: true,
			wantHour:   10,
			wantMinute: 30,
			wantField:  FieldHour,
		},
		{
			name:       "editor right increments hour",
			state:      EditorState{Active: true, Field: FieldHour, Hour: 10, Minute: 30},
			ev:         rightEv,
			wantRedraw: true,
			wantActive: true,
			wantHour:   11,
			wantMinute: 30,
			wantField:  FieldHour,
		},
		{
			name:       "editor right increments minute",
			state:      EditorState{Active: true, Field: FieldMinute, Hour: 10, Minute: 30},
			ev:         rightEv,
			wantRedraw: true,
			wantActive: true,
			wantHour:   10,
			wantMinute: 31,
			wantField:  FieldMinute,
		},
		{
			name:       "editor right on save field triggers save",
			state:      EditorState{Active: true, Field: FieldSave, Hour: 9, Minute: 5},
			ev:         rightEv,
			wantSave:   true,
			wantExit:   true,
			wantRedraw: true,
			wantHour:   9,
			wantMinute: 5,
		},
		{
			name:       "editor right on cancel field exits without save",
			state:      EditorState{Active: true, Field: FieldCancel, Hour: 9, Minute: 5},
			ev:         rightEv,
			wantExit:   true,
			wantRedraw: true,
			wantActive: false,
		},
		{
			name:       "editor both press cancels",
			state:      EditorState{Active: true, Field: FieldHour, Hour: 14, Minute: 35},
			ev:         bothEv,
			wantExit:   true,
			wantRedraw: true,
			wantActive: false,
		},
		{
			name:       "no buttons is no-op",
			state:      EditorState{Active: true, Field: FieldHour, Hour: 10, Minute: 30},
			ev:         noneEv,
			wantActive: true,
			wantHour:   10,
			wantMinute: 30,
			wantField:  FieldHour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessInput(tt.state, tt.ev, tt.onTimeScreen, window)

			if got.AttemptEntry != tt.wantAttempt {
				t.Errorf("AttemptEntry = %v, want %v", got.AttemptEntry, tt.wantAttempt)
			}
			if got.NavLeft != tt.wantNavLeft {
				t.Errorf("NavLeft = %v, want %v", got.NavLeft, tt.wantNavLeft)
			}
			if got.NavRight != tt.wantNavRight {
				t.Errorf("NavRight = %v, want %v", got.NavRight, tt.wantNavRight)
			}
			if got.ShouldSave != tt.wantSave {
				t.Errorf("ShouldSave = %v, want %v", got.ShouldSave, tt.wantSave)
			}
			if got.Exited != tt.wantExit {
				t.Errorf("Exited = %v, want %v", got.Exited, tt.wantExit)
			}
			if got.RedrawNeeded != tt.wantRedraw {
				t.Errorf("RedrawNeeded = %v, want %v", got.RedrawNeeded, tt.wantRedraw)
			}
			if got.State.Active != tt.wantActive {
				t.Errorf("State.Active = %v, want %v", got.State.Active, tt.wantActive)
			}
			if got.State.Hour != tt.wantHour {
				t.Errorf("State.Hour = %v, want %v", got.State.Hour, tt.wantHour)
			}
			if got.State.Minute != tt.wantMinute {
				t.Errorf("State.Minute = %v, want %v", got.State.Minute, tt.wantMinute)
			}
			if got.State.Field != tt.wantField {
				t.Errorf("State.Field = %v, want %v", got.State.Field, tt.wantField)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Ignored inputs
// ---------------------------------------------------------------------------

func TestIgnoredInputs(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	window := DefaultSimultaneousWindow

	bothEv := ButtonEvent{true, true, base, base}
	noneEv := ButtonEvent{false, false, time.Time{}, time.Time{}}

	tests := []struct {
		name         string
		state        EditorState
		ev           ButtonEvent
		onTimeScreen bool
	}{
		{
			name:         "simultaneous press on non-time screen is ignored",
			state:        EditorState{},
			ev:           bothEv,
			onTimeScreen: false,
		},
		{
			name:         "no buttons pressed is no-op",
			state:        EditorState{},
			ev:           noneEv,
			onTimeScreen: true,
		},
		{
			name:         "no buttons in editor is no-op",
			state:        EditorState{Active: true, Field: FieldHour, Hour: 10, Minute: 30},
			ev:           noneEv,
			onTimeScreen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessInput(tt.state, tt.ev, tt.onTimeScreen, window)

			// For all ignored inputs, no action should be taken
			if got.AttemptEntry {
				t.Error("AttemptEntry should be false for ignored input")
			}
			if got.NavLeft {
				t.Error("NavLeft should be false for ignored input")
			}
			if got.NavRight {
				t.Error("NavRight should be false for ignored input")
			}
			if got.ShouldSave {
				t.Error("ShouldSave should be false for ignored input")
			}
			if got.Exited {
				t.Error("Exited should be false for ignored input")
			}
			if got.RedrawNeeded {
				t.Error("RedrawNeeded should be false for ignored input")
			}
			// State should be unchanged
			if got.State != tt.state {
				t.Errorf("State changed: got %+v, want %+v", got.State, tt.state)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Save failure behavior
// ---------------------------------------------------------------------------

func TestHandleSaveAttempt(t *testing.T) {
	tests := []struct {
		name    string
		state   EditorState
		saveErr error
		want    EditorState
	}{
		{
			name:    "success clears editor",
			state:   EditorState{Active: false, Hour: 14, Minute: 35},
			saveErr: nil,
			want:    EditorState{},
		},
		{
			name:    "i2c error re-enters editor",
			state:   EditorState{Active: false, Hour: 14, Minute: 35},
			saveErr: errors.New("i2c bus error"),
			want:    Enter(14, 35),
		},
		{
			name:    "hardware fault re-enters editor",
			state:   EditorState{Active: false, Hour: 23, Minute: 59},
			saveErr: errors.New("device not responding"),
			want:    Enter(23, 59),
		},
		{
			name:    "success at midnight",
			state:   EditorState{Active: false, Hour: 0, Minute: 0},
			saveErr: nil,
			want:    EditorState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleSaveAttempt(tt.state, tt.saveErr)
			if got != tt.want {
				t.Errorf("HandleSaveAttempt() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Save preserves hour/minute in transition state
// ---------------------------------------------------------------------------

func TestSaveTransitionPreservesHourMinute(t *testing.T) {
	s := EditorState{Active: true, Field: FieldSave, Hour: 14, Minute: 35}
	tr := s.HandleAction(ActionRight)

	if !tr.ShouldSave {
		t.Error("expected ShouldSave")
	}
	if tr.State.Hour != 14 {
		t.Errorf("Hour = %d, want 14", tr.State.Hour)
	}
	if tr.State.Minute != 35 {
		t.Errorf("Minute = %d, want 35", tr.State.Minute)
	}
	if tr.State.Active {
		t.Error("editor should be inactive after save")
	}
}

// ---------------------------------------------------------------------------
// Navigation disabled during edit mode
// ---------------------------------------------------------------------------

func TestNavigationDisabledDuringEdit(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	window := DefaultSimultaneousWindow
	leftEv := ButtonEvent{true, false, base, time.Time{}}
	rightEv := ButtonEvent{false, true, time.Time{}, base}

	activeHour := EditorState{Active: true, Field: FieldHour, Hour: 10, Minute: 30}

	// Left press in editor should cycle field, NOT navigate
	got := ProcessInput(activeHour, leftEv, true, window)
	if got.NavLeft {
		t.Error("NavLeft should be false when editor is active")
	}
	if got.State.Field != FieldMinute {
		t.Error("left press should cycle field in editor")
	}

	// Right press in editor should increment value, NOT navigate
	got = ProcessInput(activeHour, rightEv, true, window)
	if got.NavRight {
		t.Error("NavRight should be false when editor is active")
	}
	if got.State.Hour != 11 {
		t.Error("right press should increment hour in editor")
	}
}

// ---------------------------------------------------------------------------
// Wrap behavior through ProcessInput
// ---------------------------------------------------------------------------

func TestWrapHourViaInput(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	rightEv := ButtonEvent{false, true, time.Time{}, base}

	s := EditorState{Active: true, Field: FieldHour, Hour: 23, Minute: 0}
	got := ProcessInput(s, rightEv, true, DefaultSimultaneousWindow)

	if got.State.Hour != 0 {
		t.Errorf("hour should wrap 23->0, got %d", got.State.Hour)
	}
}

func TestWrapMinuteViaInput(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	rightEv := ButtonEvent{false, true, time.Time{}, base}

	s := EditorState{Active: true, Field: FieldMinute, Hour: 10, Minute: 59}
	got := ProcessInput(s, rightEv, true, DefaultSimultaneousWindow)

	if got.State.Minute != 0 {
		t.Errorf("minute should wrap 59->0, got %d", got.State.Minute)
	}
}
