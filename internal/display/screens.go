package display

import "pico_co2/pkg/font"

// ActiveScreens returns the screens in button-navigation order.
func ActiveScreens() []Screen {
	return []Screen{
		{ID: "glance-left-mono9", Render: GlanceLeft(font.FreesansBold12, 3)},
		// {ID: "glance-left-mono9-sec5", Render: GlanceLeftSections(font.FreemonoBold9, 5, 4, 3)},
		// {ID: "glance-left-mono9-sec10", Render: GlanceLeftSections(font.FreemonoRegular9, 4, 2, 3)},
		{ID: "time", Render: RenderTime, AllowsClockEdit: true},
		// {ID: "sleep-scale", Render: RenderSleepScale},
		// {ID: "bars-with-trend", Render: RenderBarsWithTrend},
		{ID: "sparkline-co2", Render: RenderSparklineCO2},
		{ID: "sparkline-temperature", Render: RenderSparklineT},
		{ID: "sparkline-humidity", Render: RenderSparklineRH},
	}
}
