package sensors

import "pico_co2/internal/types"

type Reader interface {
	Read() ([]types.Measurement, error)
}
