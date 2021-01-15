// Package dataset handles forecasts for a single area (nordic, arctic, etc)
package dataset

import (
	"context"
	"math"
	"time"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/area"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/memory"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/pointdata"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// Dataset contains a forecast for a single area.
type Dataset struct {
	Meta pointdata.Meta

	Area    *area.Area
	readers []values.Reader
	lookups []index.Nearester
}

var downloadFunc = memory.Download

// SetDownloadFunction overrides the function to download data. It is meant for creating local tests.
func SetDownloadFunction(f func(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, hash string) (values.Reader, error)) {
	downloadFunc = f
}

// Download creates and returns a Dataset from the given specification.
func Download(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta) (*Dataset, error) {
	hashes, err := source.GetHashes(ctx, datasetMeta)
	if err != nil {
		return nil, err
	}

	var readers []values.Reader
	var lookups []index.Nearester

	for _, hash := range hashes {
		lookup, err := index.Add(ctx, source, datasetMeta, hash)
		if err != nil {
			return nil, err
		}
		lookups = append(lookups, lookup)

		reader, err := downloadFunc(ctx, source, datasetMeta, hash)
		if err != nil {
			for _, r := range readers {
				r.Close()
			}
			return nil, err
		}
		readers = append(readers, reader)
	}

	var geographicArea *area.Area
	if datasetMeta.GeographicExtent != nil {
		geographicArea, err = area.New(*datasetMeta.GeographicExtent)
		if err != nil {
			return nil, err
		}
	}

	readyTime := time.Now().UTC()
	return &Dataset{
		Meta: pointdata.Meta{
			Area:       datasetMeta.Area,
			Version:    datasetMeta.Version,
			UpdatedAt:  readyTime,
			NextUpdate: readyTime.Add(datasetMeta.TimeUntilNext),
		},
		Area:    geographicArea,
		readers: readers,
		lookups: lookups,
	}, nil
}

// Close removes all resources associated with the downloaded Dataset.
func (d *Dataset) Close() error {
	var ret error
	for _, r := range d.readers {
		if err := r.Close(); err != nil {
			ret = err
		}
	}
	return ret
}

// Read returns the best forecast for the given latitude and longitude.
func (d *Dataset) Read(latitude, longitude float32) (*pointdata.PointData, error) {
	pointData := pointdata.PointData{
		Meta: &d.Meta,
	}

	for i, r := range d.readers {
		n := d.lookups[i]
		response, err := n.Nearest(latitude, longitude)
		if err != nil {
			return nil, err
		}

		data, err := r.Read(int(response.Idx))
		if err != nil {
			return nil, err
		}
		pointData.Data = append(pointData.Data, *data)
	}
	return &pointData, nil
}

// DistanceTo returns the distance in meters from the given latitude/longitude
// to the closest point that we have data for.
func (d *Dataset) DistanceTo(latitude, longitude float32) (uint, error) {
	min := uint32(math.MaxUint32)
	for _, n := range d.lookups {
		nearest, err := n.Nearest(latitude, longitude)
		if err != nil {
			return 0, err
		}
		if nearest.Distance < min {
			min = nearest.Distance
		}
	}
	return uint(min), nil
}
