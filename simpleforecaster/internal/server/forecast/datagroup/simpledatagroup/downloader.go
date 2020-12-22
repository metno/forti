package simpledatagroup

import (
	"context"
	"encoding/binary"
	"log"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

type Downloader struct {
	store *collector.Client
}

func NewDownloader(source *collector.Client) *Downloader {
	return &Downloader{
		store: source,
	}
}

func (d *Downloader) Get(ctx context.Context, datasetMeta *collector.DatasetMeta, hash string) (*Reader, error) {

	metaCollection, err := d.store.GetHashMeta(ctx, datasetMeta, hash)
	if err != nil {
		return nil, err
	}

	data, err := d.getData(ctx, datasetMeta, hash)
	if err != nil {
		return nil, err
	}

	return &Reader{
		MetaCollection: *metaCollection,
		data:           data,
	}, nil
}

func (d *Downloader) getData(ctx context.Context, meta *collector.DatasetMeta, hash string) ([]int16, error) {
	src, err := d.store.GetData(ctx, meta, hash)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	valueCount := src.Size() / 2

	log.Printf("Download %0.1f MiB", float32(src.Size())/1024/1024)

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
