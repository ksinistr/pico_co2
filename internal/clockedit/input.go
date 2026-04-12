package clockedit

import "time"

// ButtonEvent represents raw button consumption results with timestamps.
// Timestamps record when each button was last consumed, as tracked by the
// app layer using time.Now().
type ButtonEvent struct {
	LeftPressed  bool
	RightPressed bool
	LeftTime     time.Time
	RightTime    time.Time
}

// InputResult contains the outcome of processing button input through the
// editor state machine.
type InputResult struct {
	State        EditorState
	NavLeft      bool // normal left navigation requested
	NavRight     bool // normal right navigation requested
	AttemptEntry bool // try entering edit mode (app calls Enter)
	ShouldSave   bool // save was confirmed by the user
	Exited       bool // editor exited (save or cancel)
	RedrawNeeded bool // display should be refreshed
}

// DefaultSimultaneousWindow is the recommended window for detecting
// near-simultaneous button presses.
const DefaultSimultaneousWindow = 400 * time.Millisecond

// DetectSimultaneous reports whether two button press timestamps fall within
// the given window. Both must be pressed and both timestamps must be non-zero.
func DetectSimultaneous(ev ButtonEvent, window time.Duration) bool {
	if !ev.LeftPressed || !ev.RightPressed {
		return false
	}
	if ev.LeftTime.IsZero() || ev.RightTime.IsZero() {
		return false
	}
	diff := ev.LeftTime.Sub(ev.RightTime)
	if diff < 0 {
		diff = -diff
	}
	return diff <= window
}

// ProcessInput resolves raw button input through the editor state machine.
// It handles simultaneous-press detection, editor action routing, and normal
// navigation fallback.
func ProcessInput(state EditorState, ev ButtonEvent, onTimeScreen bool, window time.Duration) InputResult {
	simultaneous := DetectSimultaneous(ev, window)
	bothInCycle := ev.LeftPressed && ev.RightPressed

	if state.Active {
		return processEditorInput(state, ev, simultaneous || bothInCycle)
	}
	return processNormalInput(state, ev, onTimeScreen, simultaneous || bothInCycle)
}

func processEditorInput(state EditorState, ev ButtonEvent, bothPressed bool) InputResult {
	var action Action
	switch {
	case bothPressed:
		action = ActionBoth
	case ev.LeftPressed:
		action = ActionLeft
	case ev.RightPressed:
		action = ActionRight
	default:
		return InputResult{State: state}
	}

	tr := state.HandleAction(action)
	return InputResult{
		State:        tr.State,
		ShouldSave:   tr.ShouldSave,
		Exited:       tr.Exit,
		RedrawNeeded: true,
	}
}

func processNormalInput(state EditorState, ev ButtonEvent, onTimeScreen bool, bothPressed bool) InputResult {
	if bothPressed {
		if onTimeScreen {
			return InputResult{
				State:        state,
				AttemptEntry: true,
				RedrawNeeded: true,
			}
		}
		// Both pressed on non-time screen: ignore both to prevent
		// accidental entry attempts.
		return InputResult{State: state}
	}

	result := InputResult{State: state}
	if ev.LeftPressed {
		result.NavLeft = true
		result.RedrawNeeded = true
	}
	if ev.RightPressed {
		result.NavRight = true
		result.RedrawNeeded = true
	}
	return result
}

// HandleSaveAttempt determines the new editor state after a save attempt.
// On success (nil error) the editor is cleared. On failure the editor is
// re-activated with the same hour/minute so the user can retry.
func HandleSaveAttempt(state EditorState, saveErr error) EditorState {
	if saveErr != nil {
		return Enter(state.Hour, state.Minute)
	}
	return EditorState{}
}
