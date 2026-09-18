package clockedit

import "time"

type ButtonEvent struct {
	LeftPressed  bool
	RightPressed bool
	LeftTime     time.Time
	RightTime    time.Time
}

type InputResult struct {
	State        EditorState
	NavLeft      bool
	NavRight     bool
	AttemptEntry bool
	ShouldSave   bool
	Exited       bool
	RedrawNeeded bool
}

// DefaultSimultaneousWindow is the recommended window for detecting
// near-simultaneous button presses.
const DefaultSimultaneousWindow = 400 * time.Millisecond

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

func processNormalInput(state EditorState, ev ButtonEvent, onTimeScreen, bothPressed bool) InputResult {
	if bothPressed {
		if onTimeScreen {
			return InputResult{
				State:        state,
				AttemptEntry: true,
				RedrawNeeded: true,
			}
		}

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

func HandleSaveAttempt(state EditorState, saveErr error) EditorState {
	if saveErr != nil {
		return Enter(state.Hour, state.Minute)
	}

	return EditorState{}
}
