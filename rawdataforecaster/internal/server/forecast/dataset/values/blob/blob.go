package blob

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/metno/forti/rawdataforecaster/pkg/fortiblob"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
)

type Reader struct {
	source fortiblob.Client
	prefix string

	datasetMeta fortiblob.DatasetMeta
	gridMeta    fortiblob.MetaCollection
	grid        string
}

func Download(ctx context.Context, source fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, grid string, config map[string]interface{}) (values.Reader, error) {
	meta, err := source.GetGridMeta(ctx, datasetMeta, grid)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s/%d/%s/", datasetMeta.Area, datasetMeta.Version, grid)

	return &Reader{
		source:      source,
		prefix:      prefix,
		datasetMeta: *datasetMeta,
		gridMeta:    *meta,
		grid:        grid,
	}, nil
}

func (r *Reader) Close() error {
	return nil
}

func (r *Reader) Read(idx int) (*values.LocationDataCollection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	reader, err := r.source.GetDataRange(ctx, &r.datasetMeta, r.grid, r.gridMeta.NumberOfPoints*idx*2, r.gridMeta.NumberOfPoints*2)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buffer := make([]int16, r.gridMeta.NumberOfPoints)
	if err := binary.Read(reader, binary.LittleEndian, &buffer); err != nil {
		return nil, err
	}

	ret := values.LocationDataCollection{
		ParameterMeta: r.gridMeta.Parameters,
		Data:          make([]float32, r.gridMeta.NumberOfPoints),
	}

	for _, meta := range r.gridMeta.Parameters {
		values.Read(&meta, ret.Data[meta.SliceFrom:], buffer[meta.SliceFrom:])
	}

	return &ret, nil
}
