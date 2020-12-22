// Package simpledatagroup handles forecasts with a single group, version and spatial resolution.
package simpledatagroup

import (
	"errors"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

// Reader serves forecast data.
type Reader struct {
	collector.MetaCollection
	data []int16
}

// Close releases any resources held by this object.
func (r *Reader) Close() error {
	return nil
}

// PointDataCollection contains forecast information for a single point
type PointDataCollection struct {
	ParameterMeta map[string]collector.ParameterMeta
	Data          []float32
}

// Read gets the forecast for the given index.
func (r *Reader) Read(idx int) (*PointDataCollection, error) {
	sliceFrom := idx * r.PointCount
	sliceTo := sliceFrom + r.PointCount

	if sliceTo >= len(r.data) {
		return nil, errors.New("out of bounds")
	}

	data := make([]float32, r.PointCount)
	for i, val := range r.data[sliceFrom:sliceTo] {
		data[i] = float32(val) / 10
	}

	return &PointDataCollection{
		ParameterMeta: r.Parameters,
		Data:          data,
	}, nil
}
