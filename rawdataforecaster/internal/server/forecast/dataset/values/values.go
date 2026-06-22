// Package values handles forecasts with a single area, version and spatial resolution.
package values

import (
	"io"

	internalformat "github.com/metno/forti-internalformat"
)

// LocationDataCollection contains forecast information for a single point
type LocationDataCollection struct {
	ParameterMeta map[string]internalformat.ParameterMeta
	Data          []float32
}

type Reader interface {
	io.Closer
	Read(idx int) (*LocationDataCollection, error)
}

// Read copies data between the two given slices, taking into consideration the scale factor.
func Read(m *internalformat.ParameterMeta, out []float32, in []int16) {
	scaleFactor := m.ScaleFactor
	if scaleFactor == 0 {
		scaleFactor = 0.1
	}

	size := len(m.Times)
	if size == 0 {
		// handle dimensions without time (eg. altitude)
		size = 1
	}
	for i := 0; i < size; i++ {
		out[i] = float32(in[i]) * scaleFactor
	}
}
