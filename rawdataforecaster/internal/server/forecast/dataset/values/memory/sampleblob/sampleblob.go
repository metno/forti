// Package sampleblob provides a sample blob reader for use in testing.
package sampleblob

import (
	"context"
	"errors"
	"io"
	"time"

	internalformat "github.com/metno/forti-internalformat"
)

func Get() internalformat.Client {
	return &client{}
}

type client struct {
}

func (c *client) Close() error {
	return nil
}

func (c *client) Latest(ctx context.Context) (map[string]int, error) {
	return map[string]int{"a": 1}, nil

}

func (c *client) GetMeta(ctx context.Context, area string, version int) (*internalformat.DatasetMeta, error) {
	if area == "a" && version == 1 {
		return &internalformat.DatasetMeta{
			Area:             "a",
			Version:          1,
			TimeUntilNext:    time.Hour,
			GeographicExtent: nil,
		}, nil
	}
	return nil, notFound
}

var notFound = errors.New("not found")

func match(area string, version int, gridid string, d *internalformat.DatasetMeta) bool {
	return area == d.Area && version == d.Version && gridid == "id_a"
}

func (c *client) GetGridInfo(ctx context.Context, d *internalformat.DatasetMeta) ([]internalformat.GridInfo, error) {
	if d.Area != "a" || d.Version != 1 {
		return nil, notFound
	}
	return []internalformat.GridInfo{
		{
			ID:          "id_a",
			RawDataSize: 100,
		},
	}, nil
}

func (c *client) GetGridMeta(ctx context.Context, d *internalformat.DatasetMeta, gridid string) (*internalformat.MetaCollection, error) {
	if !match("a", 1, "id_a", d) {
		return nil, notFound
	}
	return &internalformat.MetaCollection{
		Parameters: map[string]internalformat.ParameterMeta{
			"rr": {
				Units: "celsius",
				Times: []time.Time{
					time.Date(2021, 6, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2021, 6, 14, 2, 0, 0, 0, time.UTC),
				},
				SliceFrom:   0,
				ScaleFactor: 0.1,
			},
			"ta": {
				Units: "mm",
				Times: []time.Time{
					time.Date(2021, 6, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2021, 6, 14, 1, 0, 0, 0, time.UTC),
					time.Date(2021, 6, 14, 2, 0, 0, 0, time.UTC),
				},
				SliceFrom:   2,
				ScaleFactor: 0.1,
			},
		},
		NumberOfPoints: 5,
	}, nil
}

func (c *client) GetData(ctx context.Context, d *internalformat.DatasetMeta, gridid string) (internalformat.DataReader, error) {
	if !match("a", 1, "id_a", d) {
		return nil, notFound
	}
	var data []int16
	for i := 0; i < 100; i++ { // locations
		for j := 0; j < 5; j++ { // parameters/times
			value := int16((i * 10) + j)
			data = append(data, value)
		}
	}
	return sampleDataReader(data, 500, 2), nil
}

func (c *client) GetDataRange(ctx context.Context, d *internalformat.DatasetMeta, hash string, from, length int) (io.ReadCloser, error) {
	if !match("a", 1, "id_a", d) {
		return nil, notFound
	}
	return nil, errors.New("not implemented")
}

func (c *client) GetLatitude(ctx context.Context, d *internalformat.DatasetMeta, gridid string) (internalformat.DataReader, error) {
	if !match("a", 1, "id_a", d) {
		return nil, notFound
	}
	var data []float32
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			data = append(data, float32(j))
		}
	}
	return sampleDataReader(data, 100, 4), nil
}

func (c *client) GetLongitude(ctx context.Context, d *internalformat.DatasetMeta, gridid string) (internalformat.DataReader, error) {
	if !match("a", 1, "id_a", d) {
		return nil, notFound
	}
	var data []float32
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			data = append(data, float32(i))
		}
	}
	return sampleDataReader(data, 100, 4), nil
}
