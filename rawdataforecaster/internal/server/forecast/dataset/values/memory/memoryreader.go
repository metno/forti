package memory

import (
	"errors"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// MemoryReader serves forecast data.
type MemoryReader struct {
	fortiblob.MetaCollection
	data []int16
}

// Close releases any resources held by this object.
func (r *MemoryReader) Close() error {
	return nil
}

// Read gets the forecast for the given index.
func (r *MemoryReader) Read(idx int) (*values.LocationDataCollection, error) {
	sliceFrom := idx * r.LocationCount
	sliceTo := sliceFrom + r.LocationCount

	if sliceTo >= len(r.data) {
		return nil, errors.New("out of bounds")
	}

	data := make([]float32, r.LocationCount)
	for i, val := range r.data[sliceFrom:sliceTo] {
		data[i] = float32(val) / 10
	}

	return &values.LocationDataCollection{
		ParameterMeta: r.Parameters,
		Data:          data,
	}, nil
}
