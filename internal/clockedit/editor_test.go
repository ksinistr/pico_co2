package clockedit

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Success-path tests
// ---------------------------------------------------------------------------

func TestCanEnter(t *testing.T) {
	tests := []struct {
		name         string
		onTimeScreen bool
		wantCanEnter bool
	}{
		{"on time screen", true, true},
		{"on other screen", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanEnter(tt.onTimeScreen); got != tt.wantCanEnter {
				t.Errorf("CanEnter(%v) = %v, want %v", tt.onTimeScreen, got, tt.wantCanEnter)
			}
		})
	}
}

func TestEnterPreloadsTime(t *testing.T) {
	s := Enter(14, 35)
	if !s.Active {
		t.Fatal("expected editor to be active after Enter")
	}
	if s.Field != FieldHour {
		t.Errorf("expected initial field to be FieldHour, got %d", s.Field)
	}
	if s.Hour != 14 || s.Minute != 35 {
		t.Errorf("expected hour=14 minute=35, got hour=%d minute=%d", s.Hour, s.Minute)
	}
}

func TestMoveBetweenFields(t *testing.T) {
	tests := []struct {
		name      string
		from      EditField
		presses   int
		wantField EditField
	}{
		{"hour to minute", FieldHour, 1, FieldMinute},
		{"minute to save", FieldMinute, 1, FieldSave},
		{"save to cancel", FieldSave, 1, FieldCancel},
		{"cancel wraps to hour", FieldCancel, 1, FieldHour},
		{"full cycle", FieldHour, 4, FieldHour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := EditorState{Active: true, Field: tt.from, Hour: 10, Minute: 30}
			for i := 0; i < tt.presses; i++ {
				tr := s.HandleAction(ActionLeft)
				s = tr.State
				if tr.Exit {
					t.Fatal("unexpected exit while cycling fields")
				}
			}
			if s.Field != tt.wantField {
				t.Errorf("after %d presses from %d, got field %d, want %d",
					tt.presses, tt.from, s.Field, tt.wantField)
			}
			// Hour and minute must be preserved through field cycling.
			if s.Hour != 10 || s.Minute != 30 {
				t.Errorf("hour/minute changed during field cycle: hour=%d minute=%d", s.Hour, s.Minute)
			}
		})
	}
}

func TestWrapHour(t *testing.T) {
	tests := []struct {
		name    string
		start   int
		presses int
		want    int
	}{
		{"increment once", 10, 1, 11},
		{"wrap 23 to 0", 23, 1, 0},
		{"full 24-hour wrap", 5, 24, 5},
		{"double wrap", 0, 50, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := EditorState{Active: true, Field: FieldHour, Hour: tt.start, Minute: 0}
			for i := 0; i < tt.presses; i++ {
				tr := s.HandleAction(ActionRight)
				s = tr.State
			}
			if s.Hour != tt.want {
				t.Errorf("hour after %d presses from %d: got %d, want %d",
					tt.presses, tt.start, s.Hour, tt.want)
			}
		})
	}
}

func TestWrapMinute(t *testing.T) {
	tests := []struct {
		name    string
		start   int
		presses int
		want    int
	}{
		{"increment once", 30, 1, 31},
		{"wrap 59 to 0", 59, 1, 0},
		{"full 60-minute wrap", 15, 60, 15},
		{"double wrap", 0, 125, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := EditorState{Active: true, Field: FieldMinute, Hour: 0, Minute: tt.start}
			for i := 0; i < tt.presses; i++ {
				tr := s.HandleAction(ActionRight)
				s = tr.State
			}
			if s.Minute != tt.want {
				t.Errorf("minute after %d presses from %d: got %d, want %d",
					tt.presses, tt.start, s.Minute, tt.want)
			}
		})
	}
}

func TestSavePath(t *testing.T) {
	s := EditorState{Active: true, Field: FieldSave, Hour: 9, Minute: 5}
	tr := s.HandleAction(ActionRight)

	if !tr.ShouldSave {
		t.Error("expected ShouldSave=true on save action")
	}
	if !tr.Exit {
		t.Error("expected Exit=true on save action")
	}
	if tr.State.Active {
		t.Error("expected editor to be inactive after save")
	}
}

// ---------------------------------------------------------------------------
// Edge-case tests
// ---------------------------------------------------------------------------

func TestIgnoreEntryOnNonTimeScreen(t *testing.T) {
	if CanEnter(false) {
		t.Error("CanEnter should return false when not on time screen")
	}
}

func TestNoActionIsNoOp(t *testing.T) {
	s := EditorState{Active: true, Field: FieldHour, Hour: 14, Minute: 35}
	tr := s.HandleAction(ActionNone)
	if tr.State != s {
		t.Errorf("ActionNone changed state: got %+v, want %+v", tr.State, s)
	}
	if tr.Exit {
		t.Error("ActionNone caused unexpected exit")
	}
}

func TestInactiveEditorIgnoresAllActions(t *testing.T) {
	s := EditorState{Active: false}
	for _, a := range []Action{ActionNone, ActionLeft, ActionRight} {
		tr := s.HandleAction(a)
		if tr.State.Active {
			t.Errorf("inactive editor became active from action %d", a)
		}
		if tr.Exit {
			t.Errorf("inactive editor exited from action %d", a)
		}
	}
}

func TestCancelDiscardsChanges(t *testing.T) {
	// Start at hour=14, minute=35, change to hour=20, then cancel.
	s := Enter(14, 35)
	// Increment hour 6 times.
	for i := 0; i < 6; i++ {
		s = s.HandleAction(ActionRight).State
	}
	if s.Hour != 20 {
		t.Fatalf("expected hour=20 after increments, got %d", s.Hour)
	}
	// Move to cancel: minute -> save -> cancel (3 left presses total from hour)
	for i := 0; i < 3; i++ {
		s = s.HandleAction(ActionLeft).State
	}
	if s.Field != FieldCancel {
		t.Fatalf("expected FieldCancel, got %d", s.Field)
	}
	tr := s.HandleAction(ActionRight)
	if !tr.Exit {
		t.Error("cancel should cause exit")
	}
	if tr.ShouldSave {
		t.Error("cancel should not set ShouldSave")
	}
	if tr.State.Active {
		t.Error("editor should be inactive after cancel")
	}
}

func TestBothPressCancelsFromAnyField(t *testing.T) {
	for _, field := range []EditField{FieldHour, FieldMinute, FieldSave, FieldCancel} {
		name := [...]string{"hour", "minute", "save", "cancel"}[field]
		t.Run(name, func(t *testing.T) {
			s := EditorState{Active: true, Field: field, Hour: 12, Minute: 0}
			tr := s.HandleAction(ActionBoth)
			if !tr.Exit {
				t.Error("ActionBoth should exit from any field")
			}
			if tr.ShouldSave {
				t.Error("ActionBoth should not set ShouldSave")
			}
			if tr.State.Active {
				t.Error("editor should be inactive after ActionBoth")
			}
		})
	}
}

func TestRepeatedIncrementWraparound(t *testing.T) {
	// Incrementing hour 24 times from any value returns to the same value.
	for start := 0; start < 24; start += 5 {
		s := EditorState{Active: true, Field: FieldHour, Hour: start, Minute: 0}
		for i := 0; i < 24; i++ {
			s = s.HandleAction(ActionRight).State
		}
		if s.Hour != start {
			t.Errorf("24 increments from %d landed on %d", start, s.Hour)
		}
	}

	// Incrementing minute 60 times from any value returns to the same value.
	for start := 0; start < 60; start += 7 {
		s := EditorState{Active: true, Field: FieldMinute, Hour: 0, Minute: start}
		for i := 0; i < 60; i++ {
			s = s.HandleAction(ActionRight).State
		}
		if s.Minute != start {
			t.Errorf("60 increments from %d landed on %d", start, s.Minute)
		}
	}
}
