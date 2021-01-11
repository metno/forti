// Package modelprovider handles reading data from an object store.
package modelprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob" // support azure blobs
	_ "gocloud.dev/blob/fileblob"  // support file blobs
)

// Element enumerates all elements that should exist for a single data entry in a blob storage
type Element int

const (
	Metadata Element = iota
	Data
	Latitude
	Longitude
	AllData
)

var elementName = [4]string{
	"meta.json",
	"data",
	"latitude",
	"longitude",
}

// ToPath returns the expected blob path for the given id and element
func ToPath(id ID, element Element) string {
	return fmt.Sprintf("%s/%v/%s/%s", id.Group, id.Version, id.Param, elementName[element])
}

// Client is an interface to a blob storage model provider.
type Client struct {
	bucket *blob.Bucket
}

// FromBlob creates a Client from the given bucket
func FromBlob(bucket *blob.Bucket) *Client {
	return &Client{bucket}
}

// NewBlobClient returns a new Client, which reads from a blob storage
func NewBlobClient(bucket string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := blob.OpenBucket(ctx, bucket)
	if err != nil {
		return nil, err
	}

	return &Client{
		bucket: b,
	}, nil
}

type groupVersion struct {
	Group   string
	Version int
}

func (c *Client) keys(ctx context.Context) ([]groupVersion, error) {
	var ret []groupVersion
	re := regexp.MustCompile("complete/([a-z0-9_.]+)/([0-9]+).json")
	opt := blob.ListOptions{
		Prefix: "complete/",
	}
	li := c.bucket.List(&opt)
	for {
		item, err := li.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		elements := re.FindStringSubmatch(item.Key)
		if len(elements) == 0 {
			// no match - ignore
			continue
		}
		group := elements[1]
		version, err := strconv.Atoi(elements[2])
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s: %w", item.Key, err)
		}
		ret = append(ret, groupVersion{group, version})
	}
	return ret, nil
}

func (c *Client) Close() error {
	return c.bucket.Close()
}

func (c *Client) GetDataSet(ctx context.Context, group string, version int) (DataSet, error) {
	key := fmt.Sprintf("complete/%s/%d.json", group, version)
	r, err := c.bucket.NewReader(ctx, key, nil)
	if err != nil {
		return DataSet{}, err
	}
	defer r.Close()
	var dataset DataSet
	if err := json.NewDecoder(r).Decode(&dataset); err != nil {
		return DataSet{}, err
	}
	dataset.AvailableAt = r.ModTime()
	return dataset, nil
}

func (c *Client) Latest(ctx context.Context) ([]DataSet, error) {
	keys, err := c.keys(ctx)
	if err != nil {
		return nil, err
	}
	latest := make(map[string]int)
	for _, k := range keys {
		if k.Version > latest[k.Group] {
			latest[k.Group] = k.Version
		}
	}
	var ret []DataSet
	for group, version := range latest {
		dataset, err := c.GetDataSet(ctx, group, version)
		if err != nil {
			return nil, err
		}
		ret = append(ret, dataset)
	}

	return ret, nil
}

// Meta returns metadata about the given id.
func (c *Client) Meta(ctx context.Context, id ID) (Meta, error) {
	name := ToPath(id, Metadata)
	r, err := c.bucket.NewReader(ctx, name, nil)
	if err != nil {
		return Meta{}, err
	}
	defer r.Close()

	var meta Meta
	if err := json.NewDecoder(r).Decode(&meta); err != nil {
		return meta, fmt.Errorf("unable to download %s: %s", name, err)
	}

	return meta, nil
}

func (c *Client) CollectedMeta(ctx context.Context, group string, version int, hash string) (*CollectedMeta, error) {
	name := fmt.Sprintf("%s/%d/%s/meta.json", group, version, hash)
	r, err := c.bucket.NewReader(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var meta CollectedMeta
	if err := json.NewDecoder(r).Decode(&meta); err != nil {
		return &meta, fmt.Errorf("unable to download %s: %s", name, err)
	}

	return &meta, nil
}

// Data returns a reader for all forecast data of the given id.
func (c *Client) Data(ctx context.Context, id ID) (io.ReadCloser, error) {
	name := ToPath(id, Data)
	return c.bucket.NewReader(ctx, name, nil)
}

func (c *Client) CollectedData(ctx context.Context, group string, version int, hash string) (io.ReadCloser, error) {
	name := fmt.Sprintf("%s/%d/%s/data", group, version, hash)
	return c.bucket.NewReader(ctx, name, nil)
}

// Latitude returns a reader for all latitudes of the given id.
func (c *Client) Latitude(ctx context.Context, id ID) (io.ReadCloser, error) {
	name := ToPath(id, Latitude)
	return c.bucket.NewReader(ctx, name, nil)
}

// Longitude returns a reader for all longitudes of the given id.
func (c *Client) Longitude(ctx context.Context, id ID) (io.ReadCloser, error) {
	name := ToPath(id, Longitude)
	return c.bucket.NewReader(ctx, name, nil)
}
