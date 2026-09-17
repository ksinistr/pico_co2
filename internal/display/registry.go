package display

import (
	"pico_co2/internal/types"
)

type RenderMethod struct {
	Name string
	Fn   func(Renderer, *types.Readings)
}

var MethodRegistry = []RenderMethod{
	// {"RenderLevel", RenderLevel},
	// {"RenderCO2BarWithNums", RenderCO2BarWithNums},
	// {"RenderBars", RenderBars},
	// {"RenderTime", RenderTime},
	// {"RenderSleep", RenderSleep},
	// {"RenderSleepAdvice", RenderSleepAdvice},
	// {"RenderBarsWithLargeNums", RenderBarsWithLargeNums},
	{"RenderSleepScale", RenderSleepScale},
	{"RenderBarsWithTrend", RenderBarsWithTrend},
	// {"RenderBasic", RenderBasic},
	// {"RenderCO2Graph", RenderCO2Graph},
	// {"RenderError", RenderError},
	// {"RenderHeatIndexStatus", RenderHeatIndexStatus},
	// {"RenderLargeBar", RenderLargeBar},
	// {"RenderNums", RenderNums},
	{"RenderSparklineCO2", RenderSparklineCO2},
	// {"RenderSparklineDP", RenderSparklineDP},
	// {"RenderSparklineHI", RenderSparklineHI},
	{"RenderSparklineT", RenderSparklineT},
	{"RenderSparklineRH", RenderSparklineRH},
	// {"RenderTempHumid", RenderTempHumid},
}
