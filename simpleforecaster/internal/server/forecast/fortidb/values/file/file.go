// Package file provides an implementation of values.Reader. It is
// not meant for running in production, but for testing simpleforecaster in
// environments with little memory.
package file

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/forecast/fortidb/values"
	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

type reader struct {
	meta *collector.MetaCollection
	file *os.File
}

func Download(ctx context.Context, source *collector.Client, datasetMeta *collector.DatasetMeta, hash string) (values.Reader, error) {
	meta, err := source.GetHashMeta(ctx, datasetMeta, hash)
	if err != nil {
		return nil, err
	}

	src, err := source.GetData(ctx, datasetMeta, hash)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	sink, err := ioutil.TempFile("", "forti_")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(sink, src); err != nil {
		sink.Close()
		os.Remove(sink.Name())
		return nil, err
	}

	return &reader{
		meta: meta,
		file: sink,
	}, nil
}

func (r *reader) Close() error {
	r.file.Close()
	return os.Remove(r.file.Name())
}

func (r *reader) Read(idx int) (*values.PointDataCollection, error) {
	ret := values.PointDataCollection{
		ParameterMeta: r.meta.Parameters,
		Data:          make([]float32, r.meta.PointCount),
	}

	readFrom := idx * r.meta.PointCount * 2
	buffer := make([]byte, len(ret.Data)*2)
	if _, err := r.file.ReadAt(buffer, int64(readFrom)); err != nil {
		return nil, fmt.Errorf("b: %w", err)
	}

	bufferReader := bytes.NewReader(buffer)
	for i := range ret.Data {
		var value int16

		if err := binary.Read(bufferReader, binary.LittleEndian, &value); err != nil {
			return nil, fmt.Errorf("a: %w", err)
		}

		ret.Data[i] = float32(value) / 10
	}

	return &ret, nil
}
