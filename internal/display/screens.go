package display

import (
	"pico_co2/internal/types"
	"pico_co2/pkg/font"
)

func ActiveScreens() []Screen {
	common := []types.MeasurementType{
		types.MeasurementTypeTime,
		types.MeasurementTypeCO2,
		types.MeasurementTypeTemperature,
		types.MeasurementTypeHumidity,
	}

	return []Screen{
		{ID: "glance-left-mono9", Render: GlanceLeft(font.FreemonoRegular9, 2), Required: common},
		{ID: "time", Render: RenderTime, Required: common, AllowsClockEdit: true},
		{
			ID: "sparkline-co2", Render: RenderSparklineCO2,
			Required: []types.MeasurementType{types.MeasurementTypeCO2},
		},
		{
			ID: "sparkline-temperature", Render: RenderSparklineT,
			Required: []types.MeasurementType{types.MeasurementTypeTemperature},
		},
		{
			ID: "sparkline-humidity", Render: RenderSparklineRH,
			Required: []types.MeasurementType{types.MeasurementTypeHumidity},
		},
	}
}
