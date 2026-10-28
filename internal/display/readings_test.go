package display

import (
	"errors"
	"pico_co2/internal/types"
	"slices"
	"testing"
	"time"
)

func TestActiveScreens(t *testing.T) {
	for _, tt := range []struct {
		name     string
		readings types.Readings
	}{
		{"unavailable", types.Readings{}},
		{"warming up", types.Readings{CO2: types.Reading{State: types.ReadingWarmingUp}}},
		{"failed", types.Readings{Temperature: types.Reading{State: types.ReadingFailed}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			view := &types.ViewState{Readings: tt.readings}
			labels := make([]string, 0, 3)
			for _, cell := range glanceCells(view) {
				labels = append(labels, cell.label)
			}
			if !slices.Equal(labels, []string{"CO2", "T", "RH"}) {
				t.Fatalf("labels = %v", labels)
			}
			ids := make([]string, 0, 11)
			for _, screen := range ActiveScreens() {
				ids = append(ids, screen.ID)
			}
			want := []string{
				"glance-left-mono9", "time", "sparkline-co2", "sparkline-temperature", "sparkline-humidity",
				"readings", "bars-with-trend", "sleep-scale",
				"glance-centered-mono9", "glance-labels-mono9", "glance-left-sections-mono9",
			}
			if !slices.Equal(ids, want) {
				t.Fatalf("screens = %v", ids)
			}
			target := &countedDisplay{VirtualDisplay: NewVirtualDisplay(128, 32)}
			ActiveScreens()[0].Render(target, view)
			if target.sends != 1 {
				t.Fatal("summary was not rendered")
			}
		})
	}
}

func TestReadingPresentation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		reading types.Reading
		text    string
		message string
	}{
		{"zero is valid", types.Reading{State: types.ReadingReady}, "0c", ""},
		{"rounding", types.Reading{Value: 24.6, State: types.ReadingReady}, "25c", ""},
		{"half degree", types.Reading{Value: 24.5, State: types.ReadingReady}, "25c", ""},
		{"negative half degree", types.Reading{Value: -0.5, State: types.ReadingReady}, "-1c", ""},
		{"missing", types.Reading{}, "--", "No Temperature reading"},
		{"warming up", types.Reading{State: types.ReadingWarmingUp}, "WAIT", "Temperature warming up"},
		{"failed", types.Reading{State: types.ReadingFailed, Err: errors.New("disconnected")}, "ERR", "Temperature: disconnected"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := readingText(tt.reading, "c"); got != tt.text {
				t.Fatalf("text = %q", got)
			}
			if got := readingMessage("Temperature", tt.reading); got != tt.message {
				t.Fatalf("message = %q", got)
			}
		})
	}
}

func TestPartialReadingsAndClock(t *testing.T) {
	now := time.Date(2026, 10, 28, 14, 23, 0, 0, time.UTC)
	for _, tt := range []struct {
		name      string
		co2       types.ReadingState
		clock     types.Time
		text      string
		clockText string
	}{
		{"time unavailable", types.ReadingReady, types.Time{}, "900", "--:--"},
		{"CO2 warmup", types.ReadingWarmingUp, types.Time{Hour: 14, Minute: 23, LastRead: now}, "WAIT", "14:23"},
		{"CO2 failure", types.ReadingFailed, types.Time{}, "ERR", "--:--"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			view := &types.ViewState{
				Readings: types.Readings{
					CO2:         types.Reading{Value: 900, State: tt.co2},
					Temperature: types.Reading{Value: 24, State: types.ReadingReady},
					Humidity:    types.Reading{Value: 55, State: types.ReadingReady},
				},
				Time:    tt.clock,
				History: types.NewHistory(8, time.Minute),
			}
			cells := glanceCells(view)
			if cells[0].text != tt.text || cells[0].ready != (tt.co2 == types.ReadingReady) {
				t.Fatal("wrong CO2 presentation")
			}
			if cells[1].text != "24c" || cells[2].text != "55%" || !cells[1].ready || !cells[2].ready {
				t.Fatal("lost available climate readings")
			}
			if got := FormatClock(view.Time); got != tt.clockText {
				t.Fatalf("clock = %q", got)
			}
			screens := ActiveScreens()
			for _, screen := range screens {
				target := &countedDisplay{VirtualDisplay: NewVirtualDisplay(128, 32)}
				screen.Render(target, view)
				if target.sends != 1 {
					t.Fatalf("screen %s was not rendered", screen.ID)
				}
			}
		})
	}
}

func TestGraphReadingStates(t *testing.T) {
	for _, graph := range []struct {
		name    string
		index   int
		reading func(*types.Readings) *types.Reading
	}{
		{"CO2", 2, func(r *types.Readings) *types.Reading { return &r.CO2 }},
		{"Temperature", 3, func(r *types.Readings) *types.Reading { return &r.Temperature }},
		{"Humidity", 4, func(r *types.Readings) *types.Reading { return &r.Humidity }},
	} {
		for _, tt := range []struct {
			name    string
			state   types.ReadingState
			message string
		}{
			{"unavailable", types.ReadingUnavailable, "No " + graph.name + " reading"},
			{"warming up", types.ReadingWarmingUp, graph.name + " warming up"},
			{"failed", types.ReadingFailed, graph.name + ": disconnected"},
		} {
			t.Run(graph.name+"/"+tt.name, func(t *testing.T) {
				view := &types.ViewState{Readings: types.Readings{
					CO2:         types.Reading{State: types.ReadingReady},
					Temperature: types.Reading{State: types.ReadingReady},
					Humidity:    types.Reading{State: types.ReadingReady},
				}}
				*graph.reading(&view.Readings) = types.Reading{State: tt.state, Err: errors.New("disconnected")}
				actual := NewVirtualDisplay(128, 32)
				ActiveScreens()[graph.index].Render(actual, view)
				expected := NewVirtualDisplay(128, 32)
				RenderError(expected, tt.message)
				if !slices.Equal(actual.buffer, expected.buffer) {
					t.Fatal("graph did not display the expected reading state")
				}
			})
		}
	}
}
