package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
	"gocloud.dev/blob"
)

type Uploader struct {
	bucket *blob.Bucket
}

func New(bucket *blob.Bucket) *Uploader {
	return &Uploader{bucket}
}

func (u *Uploader) GetDataStream(ctx context.Context, area string, version int, hash string) (io.WriteCloser, error) {
	key := fmt.Sprintf("%s/%d/%s/data", area, version, hash)
	return u.bucket.NewWriter(ctx, key, nil)
}

func (u *Uploader) GetLatitudeStream(ctx context.Context, area string, version int, hash string) (io.WriteCloser, error) {
	key := fmt.Sprintf("%s/%d/%s/latitude", area, version, hash)
	return u.bucket.NewWriter(ctx, key, nil)
}

func (u *Uploader) GetLongitudeStream(ctx context.Context, area string, version int, hash string) (io.WriteCloser, error) {
	key := fmt.Sprintf("%s/%d/%s/longitude", area, version, hash)
	return u.bucket.NewWriter(ctx, key, nil)
}

func (u *Uploader) SetHashMeta(ctx context.Context, meta *collector.MetaCollection, area string, version int, hash string) error {
	key := fmt.Sprintf("%s/%d/%s/meta.json", area, version, hash)
	w, err := u.bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		return err
	}
	return w.Close()
}

func (u *Uploader) SetDatasetMeta(ctx context.Context, meta *collector.DatasetMeta) error {
	key := fmt.Sprintf("%s/%d/complete.json", meta.Area, meta.Version)
	w, err := u.bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		return err
	}
	return w.Close()
}
