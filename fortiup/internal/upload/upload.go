package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
	"gocloud.dev/blob"
)

type Uploader struct {
	bucket *blob.Bucket
}

func New(bucket *blob.Bucket) *Uploader {
	return &Uploader{bucket}
}

func (u *Uploader) GetDataStream(ctx context.Context, area string, version int, gridid string) (io.WriteCloser, error) {
	key := fmt.Sprintf("%s/%d/%s/data", area, version, gridid)
	return u.bucket.NewWriter(ctx, key, nil)
}

func (u *Uploader) GetLatitudeStream(ctx context.Context, area string, version int, gridid string) (io.WriteCloser, error) {
	key := fmt.Sprintf("%s/%d/%s/latitude", area, version, gridid)
	return u.bucket.NewWriter(ctx, key, nil)
}

func (u *Uploader) GetLongitudeStream(ctx context.Context, area string, version int, gridid string) (io.WriteCloser, error) {
	key := fmt.Sprintf("%s/%d/%s/longitude", area, version, gridid)
	return u.bucket.NewWriter(ctx, key, nil)
}

func (u *Uploader) SetGridMeta(ctx context.Context, meta *fortiblob.MetaCollection, area string, version int, gridid string) error {
	key := fmt.Sprintf("%s/%d/%s/meta.json", area, version, gridid)
	w, err := u.bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		return err
	}
	return w.Close()
}

func (u *Uploader) SetDatasetMeta(ctx context.Context, meta *fortiblob.DatasetMeta) error {
	key := fmt.Sprintf("%s/%d/complete.json", meta.Area, meta.Version)
	w, err := u.bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return u.setLatest(ctx, meta)
}

func (u *Uploader) setLatest(ctx context.Context, meta *fortiblob.DatasetMeta) error {
	key := fmt.Sprintf("latest/%s", meta.Area)
	w, err := u.bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d\n", meta.Version)

	return w.Close()
}
