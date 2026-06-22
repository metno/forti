// Package dataset handles forecasts for a single area (nordic, arctic, etc)
package dataset

import (
	"context"
	"fmt"
	"time"

	"github.com/metno/forti/rawdataforecaster/pkg/fortiblob"
	"github.com/metno/forti/rawdataforecaster/internal/server/config"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/index"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/index/grid"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/index/lookup"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
)

// Dataset contains a forecast for a single area.
type Dataset struct {
	Meta Meta

	GridLimit GridLimit
	readers   []values.Reader
	lookups   []index.Nearester
}

// GridLimit should set geographic limit for this dataset.
// Either Polygon or Distance or none is set, depending upon the type of dataset.
type GridLimit struct {
	Polygon                  *grid.Grid
	MaximumGridPointDistance int
}

type Meta struct {
	Area       string
	Version    int
	UpdatedAt  time.Time
	NextUpdate time.Time
}

// Download creates and returns a Dataset from the given specification.
func Download(ctx context.Context, source fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, cfg *config.Configuration) (*Dataset, error) {
	grids, err := source.GetGridInfo(ctx, datasetMeta)
	if err != nil {
		return nil, err
	}

	maxSizes, ok := cfg.Loader.Configuration["max_size_gib"]
	if ok {
		if err := verifySize(maxSizes.(map[string]interface{}), grids, datasetMeta); err != nil {
			return nil, err
		}
	}

	var readers []values.Reader
	var lookups []index.Nearester

	for _, grid := range grids {
		reader, err := cfg.DownloadFunction()(ctx, source, datasetMeta, grid.ID, cfg.Loader.Configuration)
		if err != nil {
			for _, r := range readers {
				r.Close()
			}
			return nil, err
		}
		readers = append(readers, reader)

		lookup, err := index.Add(ctx, source, datasetMeta, grid.ID)
		if err != nil {
			return nil, err
		}
		lookups = append(lookups, lookup)
	}

	var geographicArea *grid.Grid
	if datasetMeta.GeographicExtent != nil {
		geographicArea, err = grid.New(*datasetMeta.GeographicExtent)
		if err != nil {
			return nil, err
		}
	}

	readyTime := time.Now().UTC()
	return &Dataset{
		Meta: Meta{
			Area:       datasetMeta.Area,
			Version:    datasetMeta.Version,
			UpdatedAt:  readyTime,
			NextUpdate: readyTime.Add(datasetMeta.TimeUntilNext),
		},
		GridLimit: GridLimit{
			Polygon:                  geographicArea,
			MaximumGridPointDistance: cfg.MaximumGridPointDistance,
		},
		readers: readers,
		lookups: lookups,
	}, nil
}

func verifySize(maxSizes map[string]interface{}, grids []fortiblob.GridInfo, datasetMeta *fortiblob.DatasetMeta) error {
	maxGiB, ok := maxSizes[datasetMeta.Area]
	if ok {
		var actualSize int64
		for _, grid := range grids {
			actualSize += grid.RawDataSize
		}

		const gib float64 = 1024 * 1024 * 1024
		maxSize := int64(maxGiB.(float64) * gib)

		if actualSize > maxSize {
			return fmt.Errorf(
				"download size (%.2f GiB) is larger than allowed maximum (%.2f GiB)",
				float64(actualSize)/gib,
				float64(maxSize)/gib,
			)
		}
	}
	return nil
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
func (d *Dataset) Read(latitude, longitude float32) (*LocationData, error) {
	locationData := LocationData{
		Meta: &d.Meta,
	}

	for i, r := range d.readers {
		n := d.lookups[i]
		gridLocation, err := n.Nearest(latitude, longitude)
		if err != nil {
			return nil, err
		}

		data, err := r.Read(int(gridLocation.Idx))
		if err != nil {
			return nil, err
		}
		locationData.Data = append(locationData.Data, *data)
	}
	return &locationData, nil
}

// ClosestGridLocation returns the index GeoResponse struct for the grid location closest
// to the user requested locations.
func (d *Dataset) ClosestGridLocation(latitude, longitude float32) (*lookup.GeoResponse, error) {
	var min *lookup.GeoResponse
	for _, n := range d.lookups {
		nearest, err := n.Nearest(latitude, longitude)
		if err != nil {
			return nil, err
		}
		if min == nil || nearest.Distance < min.Distance {
			min = &nearest
		}
	}
	return min, nil
}

// ResponseHasAcceptableDistance checks if a GeoResponse is valid for a point
// that is too far away from the requested point.
func (d *Dataset) ResponseHasAcceptableDistance(georesponse *lookup.GeoResponse) bool {
	if d.GridLimit.MaximumGridPointDistance > 0 {
		if int(georesponse.Distance) > d.GridLimit.MaximumGridPointDistance {
			return false
		}
	}
	return true
}

// WithinGeographicArea tells if the given latitude/longitude is within the
// dataset's geographic area.
func (d *Dataset) WithinGeographicArea(lat, lon float32) bool {
	if d.HasPolygon() {
		location := grid.LatLon{
			Latitude:  float64(lat),
			Longitude: float64(lon),
		}
		return d.GridLimit.Polygon.Contains(location)
	}
	return true
}

func (d *Dataset) HasPolygon() bool {
	return d.GridLimit.Polygon != nil
}
