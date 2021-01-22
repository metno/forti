package blob

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

type Reader struct {
	source *fortiblob.Client
	prefix string

	datasetMeta fortiblob.DatasetMeta
	hashMeta    fortiblob.MetaCollection
	hash        string
}

func Download(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, hash string) (values.Reader, error) {
	meta, err := source.GetGridMeta(ctx, datasetMeta, hash)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s/%d/%s/", datasetMeta.Area, datasetMeta.Version, hash)

	return &Reader{
		source:      source,
		prefix:      prefix,
		datasetMeta: *datasetMeta,
		hashMeta:    *meta,
		hash:        hash,
	}, nil
}

func (r *Reader) Close() error {
	return nil
}

func (r *Reader) Read(idx int) (*values.PointDataCollection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	reader, err := r.source.GetDataRange(ctx, &r.datasetMeta, r.hash, r.hashMeta.PointCount*idx*2, r.hashMeta.PointCount*2)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buffer := make([]int16, r.hashMeta.PointCount)
	if err := binary.Read(reader, binary.LittleEndian, &buffer); err != nil {
		return nil, err
	}

	ret := values.PointDataCollection{
		ParameterMeta: r.hashMeta.Parameters,
		Data:          make([]float32, r.hashMeta.PointCount),
	}

	for i, bufData := range buffer {
		ret.Data[i] = float32(bufData) / 10
	}

	return &ret, nil
}
