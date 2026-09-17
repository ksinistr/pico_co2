package main

import (
	"fmt"
	"math"
	"os"
	"pico_co2/internal/app"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/types"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"
	"time"
)

const (
	displayWidth  int16 = 128
	displayHeight int16 = 32
)

func main() {
	if err := os.MkdirAll("images/gallery", 0755); err != nil {
		fail(err)
	}
	schedule := app.DefaultSchedule()
	readings := sampleReadings(schedule.HistoryInterval)
	if err := renderScreens(readings); err != nil {
		fail(err)
	}
	if err := renderGallery(); err != nil {
		fail(err)
	}
}

func sampleReadings(historyInterval time.Duration) *types.Readings {
	readings := types.InitReadings(
		types.DefaultHistoryCapacity,
		historyInterval,
	)
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	readings.Time.LastRead = now
	readings.Time.Hour = 14
	readings.Time.Minute = 23
	for i := range types.DefaultHistoryCapacity {
		radians := float64(i) * 2 * math.Pi / types.DefaultHistoryCapacity
		readings.AddReadingsAt(
			now.Add(time.Duration(i)*historyInterval),
			uint16(1200+300*math.Sin(radians)),
			float32(25+2*math.Sin(radians)),
			float32(50+10*math.Sin(radians)),
		)
	}
	readings.AddReadingsAt(
		now.Add(types.DefaultHistoryCapacity*historyInterval),
		1200,
		29,
		55,
	)
	return readings
}

func renderScreens(readings *types.Readings) error {
	screens := display.ActiveScreens()
	for _, screen := range screens {
		sample := *readings
		ff := func(device *display.VirtualDisplay) {
			screen.Render(device, &sample)
		}
		if err := render(screen.ID+"-normal", ff); err != nil {
			return err
		}
	}
	errorReadings := *readings
	errorReadings.Error = "Test error message for display with long text that should wrap correctly across multiple lines."
	if err := render("error", func(device *display.VirtualDisplay) { display.RenderError(device, &errorReadings) }); err != nil {
		return err
	}
	for _, edit := range []struct {
		name  string
		field types.ClockEdit
	}{
		{"edit-hour", types.ClockEdit{Active: true, Field: clockedit.FieldHour, Hour: 14, Minute: 23}},
		{"edit-minute", types.ClockEdit{Active: true, Field: clockedit.FieldMinute, Hour: 14, Minute: 23}},
		{"edit-save", types.ClockEdit{Active: true, Field: clockedit.FieldSave, Hour: 14, Minute: 23}},
		{"edit-cancel", types.ClockEdit{Active: true, Field: clockedit.FieldCancel, Hour: 14, Minute: 23}},
	} {
		sample := *readings
		sample.ClockEdit = edit.field
		if err := render("time-"+edit.name, func(device *display.VirtualDisplay) { display.RenderTime(device, &sample) }); err != nil {
			return err
		}
	}
	return nil
}

func renderGallery() error {
	values := []float32{20, 24, 28, 32, 36, 40}
	for _, value := range values {
		name := fmt.Sprintf("gallery/thermal-%02d", int(value))
		if err := render(name, func(device *display.VirtualDisplay) {
			widget.ThermalScale(device, 14, 3, value, widget.ThermalRange{Min: 24, Warm: 28, Hot: 32, Max: 36})
		}); err != nil {
			return err
		}
	}
	for _, value := range []int16{-3, 0, 3} {
		name := fmt.Sprintf("gallery/bar-%d", value)
		if err := render(name, func(device *display.VirtualDisplay) {
			widget.TwoSideBar(device, font.New(device, font.ProggySZ8), 0, 8, value, "TEST", 3, 3)
		}); err != nil {
			return err
		}
	}
	for _, direction := range []widget.TrendDirection{widget.TrendRising, widget.TrendStable, widget.TrendFalling, widget.TrendUnknown} {
		name := fmt.Sprintf("gallery/trend-%d", direction)
		if err := render(name, func(device *display.VirtualDisplay) { widget.Trend(device, 64, 16, direction) }); err != nil {
			return err
		}
	}
	if err := render("gallery/square-bar", func(device *display.VirtualDisplay) { widget.SquareBar(device, 0, 12, 3) }); err != nil {
		return err
	}
	if err := render("gallery/square-bar-empty", func(device *display.VirtualDisplay) { widget.SquareBar(device, 0, 12, 0) }); err != nil {
		return err
	}
	if err := render("gallery/square-bar-full", func(device *display.VirtualDisplay) { widget.SquareBar(device, 0, 12, 4) }); err != nil {
		return err
	}
	if err := render("gallery/vertical-bar", func(device *display.VirtualDisplay) { widget.VerticalBar(device, 60, 8, 3, 4) }); err != nil {
		return err
	}
	if err := render("gallery/vertical-bar-empty", func(device *display.VirtualDisplay) { widget.VerticalBar(device, 60, 8, 0, 4) }); err != nil {
		return err
	}
	if err := render("gallery/vertical-bar-full", func(device *display.VirtualDisplay) { widget.VerticalBar(device, 60, 8, 4, 4) }); err != nil {
		return err
	}
	for _, text := range []string{"short", "a deliberately long message that wraps over the display width"} {
		name := "gallery/text-" + fmt.Sprintf("%d", len(text))
		if err := render(name, func(device *display.VirtualDisplay) {
			layout.LongText(device, font.New(device, font.ProggySZ8), 0, 0, text)
		}); err != nil {
			return err
		}
	}
	if err := render("gallery/row", func(device *display.VirtualDisplay) {
		face := font.New(device, font.ProggySZ8)
		layout.Three(device, face, 12, "LEFT", "MID", "RIGHT")
	}); err != nil {
		return err
	}
	if err := render("gallery/sparkline", func(device *display.VirtualDisplay) {
		widget.Sparkline(device, 0, 8, 128, 21, []int16{0, 10, 30, 10, 50, 20, 0})
	}); err != nil {
		return err
	}
	return render(
		"gallery/sparkline-empty",
		func(device *display.VirtualDisplay) {
			widget.Sparkline(device, 0, 8, 128, 21, nil)
		},
	)
}

func render(name string, draw func(*display.VirtualDisplay)) error {
	device := display.NewVirtualDisplay(displayWidth, displayHeight)
	device.Clear()
	draw(device)
	return device.SavePNG("images/" + name + ".png")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
