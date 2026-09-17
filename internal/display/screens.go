package display

// ActiveScreens returns the screens in button-navigation order.
func ActiveScreens() []Screen {
	return []Screen{
		{ID: "time", Render: RenderTime, AllowsClockEdit: true},
		{ID: "sleep-scale", Render: RenderSleepScale},
		{ID: "bars-with-trend", Render: RenderBarsWithTrend},
		{ID: "sparkline-co2", Render: RenderSparklineCO2},
		{ID: "sparkline-temperature", Render: RenderSparklineT},
		{ID: "sparkline-humidity", Render: RenderSparklineRH},
	}
}
