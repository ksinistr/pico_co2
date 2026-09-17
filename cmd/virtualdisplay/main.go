package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"pico_co2/internal/display"
	"pico_co2/internal/types"
	"pico_co2/pkg/font"
	"pico_co2/pkg/layout"
	"pico_co2/pkg/widget"
)

const (
	queueCapacity       = 480
	displayWidth  int16 = 128
	displayHeight int16 = 32
)

func main() {
	if err := os.MkdirAll("images/gallery", 0755); err != nil {
		fail(err)
	}
	readings := sampleReadings()
	if err := renderScreens(readings); err != nil {
		fail(err)
	}
	if err := renderGallery(); err != nil {
		fail(err)
	}
}

func sampleReadings() *types.Readings {
	readings := types.InitReadings(queueCapacity)
	now := time.Date(2026, 9, 17, 14, 23, 0, 0, time.UTC)
	readings.Time.LastRead = now
	readings.Time.Hour = 14
	readings.Time.Minute = 23
	for i := 0; i < queueCapacity; i++ {
		radians := float64(i) * 2 * math.Pi / queueCapacity
		readings.AddReadingsAt(
			now.Add(time.Duration(i)*time.Minute),
			uint16(1200+300*math.Sin(radians)),
			float32(25+2*math.Sin(radians)),
			float32(50+10*math.Sin(radians)),
		)
	}
	readings.AddReadingsAt(now.Add(queueCapacity*time.Minute), 1200, 26, 55)
	return readings
}

func renderScreens(readings *types.Readings) error {
	screens := display.ActiveScreens()
	for _, state := range []struct {
		name  string
		edit  types.ClockEdit
		error string
	}{
		{name: "normal"},
		{name: "error", error: "Test error message for display with long text that should wrap correctly across multiple lines."},
	} {
		for _, screen := range screens {
			copy := *readings
			copy.Error = state.error
			if copy.Error != "" {
				if err := render(screenImageName(screen)+"-"+state.name, func(device *display.VirtualDisplay) { display.RenderError(device, &copy) }); err != nil {
					return err
				}
			} else if err := render(screenImageName(screen)+"-"+state.name, func(device *display.VirtualDisplay) { screen.Render(device, &copy) }); err != nil {
				return err
			}
		}
	}
	for _, edit := range []struct {
		name  string
		field int
	}{
		{"edit-hour", types.EditFieldHour}, {"edit-minute", types.EditFieldMinute}, {"edit-save", types.EditFieldSave}, {"edit-cancel", types.EditFieldCancel},
	} {
		copy := *readings
		copy.ClockEdit = types.ClockEdit{Active: true, Field: edit.field, Hour: 14, Minute: 23}
		if err := render("RenderTime-"+edit.name, func(device *display.VirtualDisplay) { display.RenderTime(device, &copy) }); err != nil {
			return err
		}
	}
	return nil
}

func renderGallery() error {
	values := []float32{20, 24, 28, 32, 36, 40}
	for _, value := range values {
		name := fmt.Sprintf("gallery/thermal-%02d", int(value))
		if err := render(name, func(device *display.VirtualDisplay) { widget.ThermalScale(device, 14, 5, value, 24, 28, 32, 36) }); err != nil {
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
	for _, direction := range []widget.TrendDirection{widget.TrendRising, widget.TrendStable, widget.TrendFalling} {
		name := fmt.Sprintf("gallery/trend-%d", direction)
		if err := render(name, func(device *display.VirtualDisplay) { widget.Trend(device, 64, 16, direction) }); err != nil {
			return err
		}
	}
	if err := render("gallery/square-bar", func(device *display.VirtualDisplay) { widget.SquareBar(device, 0, 12, 3) }); err != nil {
		return err
	}
	if err := render("gallery/vertical-bar", func(device *display.VirtualDisplay) { widget.VerticalBar(device, 60, 8, 3, 4) }); err != nil {
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
	return render("gallery/sparkline", func(device *display.VirtualDisplay) {
		widget.Sparkline(device, 0, 8, 128, 21, []int16{0, 10, 30, 10, 50, 20, 0})
	})
}

func screenImageName(screen display.Screen) string {
	switch screen.ID {
	case "time":
		return "RenderTime"
	case "sleep-scale":
		return "RenderSleepScale"
	case "bars-with-trend":
		return "RenderBarsWithTrend"
	case "sparkline-co2":
		return "RenderSparklineCO2"
	case "sparkline-temperature":
		return "RenderSparklineT"
	case "sparkline-humidity":
		return "RenderSparklineRH"
	default:
		return screen.ID
	}
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
