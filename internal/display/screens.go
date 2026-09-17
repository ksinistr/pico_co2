package display

import "pico_co2/pkg/font"

// ActiveScreens returns the screens in button-navigation order.
func ActiveScreens() []Screen {
	return []Screen{
		{ID: "glance-left-mono9", Render: GlanceLeft(font.FreemonoRegular9, 6)},
		{ID: "glance-centered-mono12", Render: GlanceCentered(font.FreemonoRegular12, 4)},
		{ID: "time", Render: RenderTime, AllowsClockEdit: true},
		{ID: "sleep-scale", Render: RenderSleepScale},
		{ID: "bars-with-trend", Render: RenderBarsWithTrend},
		{ID: "sparkline-co2", Render: RenderSparklineCO2},
		{ID: "sparkline-temperature", Render: RenderSparklineT},
		{ID: "sparkline-humidity", Render: RenderSparklineRH},
	}
}
