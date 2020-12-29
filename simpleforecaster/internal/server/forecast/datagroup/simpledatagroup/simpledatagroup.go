// Package simpledatagroup handles forecasts with a single group, version and spatial resolution.
package simpledatagroup

import (
	"io"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

// PointDataCollection contains forecast information for a single point
type PointDataCollection struct {
	ParameterMeta map[string]collector.ParameterMeta
	Data          []float32
}

type Reader interface {
	io.Closer
	Read(idx int) (*PointDataCollection, error)
}
