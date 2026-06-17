package memory

import (
	"errors"

	"github.com/metno/forti/fortiup/pkg/fortiblob"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
)

// MemoryReader serves forecast data.
type MemoryReader struct {
	fortiblob.MetaCollection
	mad *manuallyAllocatedData
}

// Close releases any resources held by this object.
func (r *MemoryReader) Close() error {
	r.mad.Free()
	return nil
}

// Read gets the forecast for the given index.
func (r *MemoryReader) Read(idx int) (*values.LocationDataCollection, error) {
	sliceFrom := idx * r.NumberOfPoints
	sliceTo := sliceFrom + r.NumberOfPoints

	if sliceTo > len(r.mad.Values) {
		return nil, errors.New("out of bounds")
	}

	in := r.mad.Values[sliceFrom:sliceTo]
	out := make([]float32, len(in))

	for _, meta := range r.Parameters {
		values.Read(&meta, out[meta.SliceFrom:], in[meta.SliceFrom:])
	}

	return &values.LocationDataCollection{
		ParameterMeta: r.Parameters,
		Data:          out,
	}, nil
}
