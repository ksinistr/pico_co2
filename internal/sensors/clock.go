package sensors

import (
	"fmt"
	"pico_co2/internal/types"
	"time"
)

type (
	ClockReader interface{ ReadTime() (time.Time, error) }
	Clock       struct{ Source ClockReader }
)

func (c Clock) Read() ([]types.Measurement, error) {
	value, err := c.Source.ReadTime()
	if err != nil {
		return nil, fmt.Errorf("read time: %w", err)
	}

	return []types.Measurement{{Type: types.MeasurementTypeTime, Value: float64(value.Unix())}}, nil
}
