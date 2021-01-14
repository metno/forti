package fortiblob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
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
	ret := make(map[string]int)
	it := c.bucket.List(nil)
	for {
		lo, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return ret, nil
			}
			return nil, err
		}
		if strings.HasSuffix(lo.Key, "/complete.json") {
			elements := strings.Split(lo.Key, "/")
			if len(elements) == 3 {
				area := elements[0]
				version, err := strconv.Atoi(elements[1])
				if err != nil {
					log.Printf("enountered unexpected key in store: %s", lo.Key)
					continue
				}
				if version > ret[area] {
					ret[area] = version
				}
			}
		}
	}
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

func (c *Client) GetHashes(ctx context.Context, d *DatasetMeta) ([]string, error) {
	var hashes []string
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

			hashes = append(hashes, elements[2])
		}
	}

	return hashes, nil
}

func (c *Client) GetHashMeta(ctx context.Context, d *DatasetMeta, hash string) (*MetaCollection, error) {
	path := fmt.Sprintf("%s/%d/%s/meta.json", d.Area, d.Version, hash)

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

func (c *Client) GetData(ctx context.Context, d *DatasetMeta, hash string) (DataReader, error) {
	return c.getStream(ctx, fmt.Sprintf("%s/%d/%s/data", d.Area, d.Version, hash))
}

func (c *Client) GetLatitude(ctx context.Context, d *DatasetMeta, hash string) (DataReader, error) {
	return c.getStream(ctx, fmt.Sprintf("%s/%d/%s/latitude", d.Area, d.Version, hash))
}

func (c *Client) GetLongitude(ctx context.Context, d *DatasetMeta, hash string) (DataReader, error) {
	return c.getStream(ctx, fmt.Sprintf("%s/%d/%s/longitude", d.Area, d.Version, hash))
}

func (c *Client) getStream(ctx context.Context, path string) (DataReader, error) {
	r, err := c.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}
