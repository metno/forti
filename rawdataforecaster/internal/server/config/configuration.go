package config

import (
	"context"

	internalformat "github.com/metno/forti-internalformat"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values/blob"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values/memory"
)

type DownloadFunction func(ctx context.Context, source internalformat.Client, datasetMeta *internalformat.DatasetMeta, gridid string, config map[string]interface{}) (values.Reader, error)

type Configuration struct {
	Source                   DataSource `json:"source"`
	Areas                    []string   `json:"areas"`
	Loader                   Loader     `json:"loader"`
	MaximumGridPointDistance int        `json:"maximum_gridpoint_distance"` // in meters
}

func (c *Configuration) DownloadFunction() DownloadFunction {
	switch c.Loader.Type {
	case "memory":
		return memory.Download
	case "blob":
		return blob.Download
	}
	panic("no such downloader: " + c.Loader.Type)
}

type DataSource struct {
	Bucket string `json:"bucket"`
}

type Loader struct {
	Type          string                 `json:"type"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}
