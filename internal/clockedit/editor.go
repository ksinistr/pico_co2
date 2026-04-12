// Package clockedit defines the interaction model and state transitions
// for setting the DS3231 clock using the two touch buttons.
//
// User flow (code-level specification):
//
//  1. Entry: clock setup can only be entered from the time screen (the first
//     display in the registry). The user must press both touch buttons within
//     a ~400 ms window. Any narrower or wider combination is ignored.
//
//  2. Editing: once active, normal left/right display navigation is disabled.
//     - Left button cycles the cursor: hour -> minute -> save -> cancel -> hour
//     - Right button either increments the selected numeric field (hour wraps
//     0-23, minute wraps 0-59) or activates the selected action.
//
//  3. Save: selecting "save" and pressing right writes the pending hour and
//     minute to the DS3231, preserving the current date and resetting seconds
//     to zero. The editor exits and normal navigation resumes.
//
//  4. Cancel: selecting "cancel" and pressing right (or pressing both buttons
//     again) discards all changes and returns to normal navigation.
//
// State boundaries:
//   - This package owns pure state types and transition logic only.
//   - RTC reads/writes, display rendering, and button polling remain in the
//     app, display, and button packages respectively.
package clockedit

// EditField identifies which field is currently selected in the editor.
type EditField int

const (
	FieldHour   EditField = iota
	FieldMinute
	FieldSave
	FieldCancel
)

// Action represents a button event interpreted by the editor.
type Action int

const (
	ActionNone  Action = iota // no button event
	ActionLeft                 // left button pressed: cycle selected field
	ActionRight                // right button pressed: increment or activate
	ActionBoth                 // both buttons pressed: toggle editor
)

// EditorState holds the pending clock values and cursor position.
type EditorState struct {
	Active bool
	Field  EditField
	Hour   int
	Minute int
}

// Transition is the result of applying an action to an editor state.
type Transition struct {
	State      EditorState
	ShouldSave bool // true when the user confirmed save
	Exit       bool // true when the editor should close (save or cancel)
}

// Enter creates a new active editor state preloaded with the given time.
func Enter(hour, minute int) EditorState {
	return EditorState{
		Active: true,
		Field:  FieldHour,
		Hour:   hour,
		Minute: minute,
	}
}

// CanEnter reports whether the editor can be activated from the current context.
// Clock setup is only available from the time screen.
func CanEnter(isOnTimeScreen bool) bool {
	return isOnTimeScreen
}

// HandleAction applies an action to the current state and returns the resulting
// transition. When the editor is inactive, all actions are no-ops except
// ActionBoth which is handled by the caller after checking CanEnter.
func (s EditorState) HandleAction(a Action) Transition {
	if !s.Active {
		return Transition{State: s}
	}

	switch a {
	case ActionLeft:
		return Transition{State: s.moveField()}
	case ActionRight:
		return s.applyRight()
	case ActionBoth:
		return Transition{
			State: EditorState{Active: false},
			Exit:  true,
		}
	default:
		return Transition{State: s}
	}
}

// moveField advances the cursor to the next field in the cycle:
// hour -> minute -> save -> cancel -> hour.
func (s EditorState) moveField() EditorState {
	return EditorState{
		Active: s.Active,
		Field:  EditField((int(s.Field) + 1) % 4),
		Hour:   s.Hour,
		Minute: s.Minute,
	}
}

// applyRight handles the right button press depending on the selected field.
func (s EditorState) applyRight() Transition {
	switch s.Field {
	case FieldHour:
		return Transition{State: EditorState{
			Active: s.Active,
			Field:  s.Field,
			Hour:   (s.Hour + 1) % 24,
			Minute: s.Minute,
		}}
	case FieldMinute:
		return Transition{State: EditorState{
			Active: s.Active,
			Field:  s.Field,
			Hour:   s.Hour,
			Minute: (s.Minute + 1) % 60,
		}}
	case FieldSave:
		return Transition{
			State:      EditorState{Active: false, Hour: s.Hour, Minute: s.Minute},
			ShouldSave: true,
			Exit:       true,
		}
	case FieldCancel:
		return Transition{
			State: EditorState{Active: false},
			Exit:  true,
		}
	default:
		return Transition{State: s}
	}
}
