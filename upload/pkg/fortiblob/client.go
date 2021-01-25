package fortiblob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gocloud.dev/blob"
)

type Client struct {
	bucket *blob.Bucket
}

func NewClient(connectURL string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bucket, err := blob.OpenBucket(ctx, connectURL)
	if err != nil {
		return nil, err
	}

	return NewClientFromBucket(bucket), nil
}

func NewClientFromBucket(bucket *blob.Bucket) *Client {
	return &Client{bucket}
}

func (c *Client) Close() error {
	return c.bucket.Close()
}

func (c *Client) Latest(ctx context.Context) (map[string]int, error) {
	prefix := "latest/"

	ret := make(map[string]int)
	it := c.bucket.List(&blob.ListOptions{
		Prefix: prefix,
	})
	for {
		lo, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		area := strings.TrimPrefix(lo.Key, prefix)

		r, err := c.bucket.NewReader(ctx, lo.Key, nil)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		var version int
		if _, err := fmt.Fscanf(r, "%d", &version); err != nil {
			return nil, err
		}
		ret[area] = version
	}
	return ret, nil
}

func (c *Client) GetMeta(ctx context.Context, area string, version int) (*DatasetMeta, error) {
	path := fmt.Sprintf("%s/%d/complete.json", area, version)

	r, err := c.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var ret DatasetMeta
	if err := json.NewDecoder(r).Decode(&ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

type GridInfo struct {
	ID          string
	RawDataSize int64
}

func (c *Client) GetGridInfo(ctx context.Context, d *DatasetMeta) ([]GridInfo, error) {
	var gridids []GridInfo
	it := c.bucket.List(
		&blob.ListOptions{
			Prefix:    fmt.Sprintf("%s/%d/", d.Area, d.Version),
			Delimiter: "/",
		},
	)
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if obj.IsDir {
			elements := strings.Split(obj.Key, "/")
			if len(elements) != 4 {
				continue
			}

			attr, err := c.bucket.Attributes(ctx, obj.Key+"data")
			if err != nil {
				return nil, err
			}

			gridids = append(gridids,
				GridInfo{
					ID:          elements[2],
					RawDataSize: attr.Size,
				},
			)
		}
	}

	return gridids, nil
}

func (c *Client) GetGridMeta(ctx context.Context, d *DatasetMeta, gridid string) (*MetaCollection, error) {
	path := fmt.Sprintf("%s/%d/%s/meta.json", d.Area, d.Version, gridid)

	r, err := c.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var ret MetaCollection
	if err := json.NewDecoder(r).Decode(&ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

type DataReader interface {
	io.ReadCloser
	Size() int64
}

func (c *Client) GetData(ctx context.Context, d *DatasetMeta, gridid string) (DataReader, error) {
	return c.getStream(ctx, fmt.Sprintf("%s/%d/%s/data", d.Area, d.Version, gridid))
}

func (c *Client) GetDataRange(ctx context.Context, d *DatasetMeta, hash string, from, length int) (io.ReadCloser, error) {
	path := fmt.Sprintf("%s/%d/%s/data", d.Area, d.Version, hash)
	r, err := c.bucket.NewRangeReader(ctx, path, int64(from), int64(length), nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) GetLatitude(ctx context.Context, d *DatasetMeta, gridid string) (DataReader, error) {
	return c.getStream(ctx, fmt.Sprintf("%s/%d/%s/latitude", d.Area, d.Version, gridid))
}

func (c *Client) GetLongitude(ctx context.Context, d *DatasetMeta, gridid string) (DataReader, error) {
	return c.getStream(ctx, fmt.Sprintf("%s/%d/%s/longitude", d.Area, d.Version, gridid))
}

func (c *Client) getStream(ctx context.Context, path string) (DataReader, error) {
	r, err := c.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}
