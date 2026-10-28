package display

import "pico_co2/pkg/font"

func ActiveScreens() []Screen {
	return []Screen{
		{ID: "easyread", Render: EasyRead},
		{ID: "glance-left-mono9", Render: GlanceLeft(font.FreemonoRegular9, 2)},
		{ID: "time", Render: RenderTime, AllowsClockEdit: true},
		{ID: "sparkline-co2", Render: RenderSparklineCO2},
		{ID: "sparkline-temperature", Render: RenderSparklineT},
		{ID: "sparkline-humidity", Render: RenderSparklineRH},
		{ID: "bars-with-trend", Render: RenderBarsWithTrend},
		{ID: "sleep-scale", Render: RenderSleepScale},
		{ID: "glance-centered-mono9", Render: GlanceCentered(font.FreemonoRegular9, 2)},
		{ID: "glance-labels-mono9", Render: GlanceLabels(font.FreemonoRegular9, 2)},
		{ID: "glance-left-sections-mono9", Render: GlanceLeftSections(font.FreemonoRegular9, 2, 2, 1)},
	}
}
