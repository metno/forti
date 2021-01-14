// Package values handles forecasts with a single area, version and spatial resolution.
package values

import (
	"io"

	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// PointDataCollection contains forecast information for a single point
type PointDataCollection struct {
	ParameterMeta map[string]fortiblob.ParameterMeta
	Data          []float32
}

type Reader interface {
	io.Closer
	Read(idx int) (*PointDataCollection, error)
}
