package main

import (
	"math"
	"os"
	"pico_co2/internal/display"
	"pico_co2/internal/types"
	"time"
)

const (
	queueCapacity       = 480 // Number of readings to keep in memory
	displayWidth  int16 = 128
	displayHeight int16 = 32
)

func simulateSensor(x uint8) uint16 {
	radians := float32(x) * 2 * math.Pi / queueCapacity
	y := 1200 + 800*float32(math.Sin(float64(radians)))
	return uint16(y)
}

func main() {
	vd := display.NewVirtualDisplay(displayWidth, displayHeight)

	testReadings := types.InitReadings(queueCapacity)
	testReadings.Time.LastRead = time.Now()
	testReadings.Time.Hour = 14
	testReadings.Time.Minute = 23
	testReadings.FirstReadingAt = testReadings.FirstReadingAt.Add(
		-3 * time.Minute,
	)

	countMeasurements := queueCapacity
	for i := range countMeasurements {
		testReadings.History.AddedAt = testReadings.FirstReadingAt.Add(
			-2 * time.Minute,
		)
		// use formula to generate graph data with increasing and decreasing values
		co2 := simulateSensor(uint8(i))
		temperature := 22.5 + float64(i)/10.0
		humidity := 45.0 + float64(i)/10.0
		testReadings.AddReadings(
			uint16(co2),
			float32(temperature),
			float32(humidity),
		)
	}

	testCases := []struct {
		name     string
		mode     string
		readings func(*types.Readings) *types.Readings
		displays []display.RenderMethod
	}{
		{
			name: "normal",
			readings: func(r *types.Readings) *types.Readings {
				r.Error = ""
				return r
			},
			displays: display.MethodRegistry,
		},
		{
			name: "error",
			readings: func(r *types.Readings) *types.Readings {
				r.Error = "Test error message for display with long text that should wrap correctly across multiple lines."
				return r
			},
			displays: display.MethodRegistry,
		},
		{
			name: "edit-hour",
			readings: func(r *types.Readings) *types.Readings {
				r.Error = ""
				r.ClockEdit = types.ClockEdit{
					Active: true,
					Field:  types.EditFieldHour,
					Hour:   14,
					Minute: 23,
				}
				return r
			},
			displays: []display.RenderMethod{
				{Name: "RenderTime", Fn: display.RenderTime},
			},
		},
		{
			name: "edit-minute",
			readings: func(r *types.Readings) *types.Readings {
				r.Error = ""
				r.ClockEdit = types.ClockEdit{
					Active: true,
					Field:  types.EditFieldMinute,
					Hour:   14,
					Minute: 23,
				}
				return r
			},
			displays: []display.RenderMethod{
				{Name: "RenderTime", Fn: display.RenderTime},
			},
		},
		{
			name: "edit-save",
			readings: func(r *types.Readings) *types.Readings {
				r.Error = ""
				r.ClockEdit = types.ClockEdit{
					Active: true,
					Field:  types.EditFieldSave,
					Hour:   14,
					Minute: 23,
				}
				return r
			},
			displays: []display.RenderMethod{
				{Name: "RenderTime", Fn: display.RenderTime},
			},
		},
		{
			name: "edit-cancel",
			readings: func(r *types.Readings) *types.Readings {
				r.Error = ""
				r.ClockEdit = types.ClockEdit{
					Active: true,
					Field:  types.EditFieldCancel,
					Hour:   14,
					Minute: 23,
				}
				return r
			},
			displays: []display.RenderMethod{
				{Name: "RenderTime", Fn: display.RenderTime},
			},
		},
	}

	for _, tc := range testCases {
		vd.Clear()
		for _, method := range tc.displays {
			os.MkdirAll("images", 0755)
			fileName := "images/" + method.Name + "-" + tc.name + ".png"
			method.Fn(vd, tc.readings(testReadings))
			vd.SavePNG(fileName)
		}
	}
}
