package memory

import (
	"context"
	"encoding/binary"
	"log"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func Download(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, gridid string, config map[string]interface{}) (values.Reader, error) {
	d := downloader{
		store: source,
	}
	return d.Get(ctx, datasetMeta, gridid)
}

type downloader struct {
	store *fortiblob.Client
}

func newDownloader(source *fortiblob.Client) *downloader {
	return &downloader{
		store: source,
	}
}

func (d *downloader) Get(ctx context.Context, datasetMeta *fortiblob.DatasetMeta, gridid string) (values.Reader, error) {

	metaCollection, err := d.store.GetGridMeta(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}

	data, err := d.getData(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}

	return &MemoryReader{
		MetaCollection: *metaCollection,
		data:           data,
	}, nil
}

func (d *downloader) getData(ctx context.Context, meta *fortiblob.DatasetMeta, gridid string) ([]int16, error) {
	src, err := d.store.GetData(ctx, meta, gridid)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	valueCount := src.Size() / 2

	log.Printf("Download %0.2f GiB", float64(src.Size())/(1024*1024*1024))

	// We return an array with an extra entry.
	// First reserve the extra cap, and when returning we increase len with 1.
	// This is because we want to be able to slice like this for the last element:
	// ret[idx:len(ret)], which would otherwise fail.
	ret := make([]int16, valueCount, valueCount+1)

	chunkSize := 1024 * 1024
	for i := 0; i < int(valueCount); i += chunkSize {
		var part []int16
		if i+chunkSize > int(valueCount) {
			part = ret[i:]
		} else {
			part = ret[i : i+chunkSize]
		}
		if err := binary.Read(src, binary.LittleEndian, &part); err != nil {
			return nil, err
		}
	}

	return append(ret, 0), nil
}
