package memory

import (
	"errors"

	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/forecast/datagroup/values"
	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

// MemoryReader serves forecast data.
type MemoryReader struct {
	collector.MetaCollection
	data []int16
}

// Close releases any resources held by this object.
func (r *MemoryReader) Close() error {
	return nil
}

// Read gets the forecast for the given index.
func (r *MemoryReader) Read(idx int) (*values.PointDataCollection, error) {
	sliceFrom := idx * r.PointCount
	sliceTo := sliceFrom + r.PointCount

	if sliceTo >= len(r.data) {
		return nil, errors.New("out of bounds")
	}

	data := make([]float32, r.PointCount)
	for i, val := range r.data[sliceFrom:sliceTo] {
		data[i] = float32(val) / 10
	}

	return &values.PointDataCollection{
		ParameterMeta: r.Parameters,
		Data:          data,
	}, nil
}
