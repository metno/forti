package memory

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"log"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func Download(ctx context.Context, source fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, gridid string, config map[string]interface{}) (values.Reader, error) {
	d := downloader{
		store: source,
	}
	return d.Get(ctx, datasetMeta, gridid)
}

type downloader struct {
	store fortiblob.Client
}

func newDownloader(source fortiblob.Client) *downloader {
	return &downloader{
		store: source,
	}
}

func (d *downloader) Get(ctx context.Context, datasetMeta *fortiblob.DatasetMeta, gridid string) (values.Reader, error) {

	metaCollection, err := d.store.GetGridMeta(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}

	mad, err := d.getMad(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}

	return &MemoryReader{
		MetaCollection: *metaCollection,
		mad:            mad,
	}, nil
}

func (d *downloader) getMad(ctx context.Context, meta *fortiblob.DatasetMeta, gridid string) (*manuallyAllocatedData, error) {
	src, err := d.store.GetData(ctx, meta, gridid)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	log.Printf("Download %0.2f GiB from %s/%d/%s", float64(src.Size())/(1024*1024*1024), meta.Area, meta.Version, gridid)

	return readData(src)
}

func readData(src fortiblob.DataReader) (*manuallyAllocatedData, error) {
	valueSize := int64(2) // sizeof(int16)
	mad := allocate(int(src.Size() / valueSize))

	r := bufio.NewReader(src)
	buf := make([]byte, valueSize)
	for i := range mad.Values {
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		mad.Values[i] = int16(binary.LittleEndian.Uint16(buf))
	}
	return mad, nil
}
