package config

import (
	"context"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/blob"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/file"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values/memory"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

type DownloadFunction func(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, gridid string, config map[string]interface{}) (values.Reader, error)

type Configuration struct {
	Source DataSource `json:"source"`
	Areas  []string   `json:"areas"`
	Loader Loader     `json:"loader"`
}

func (c *Configuration) DownloadFunction() DownloadFunction {
	switch c.Loader.Type {
	case "memory":
		return memory.Download
	case "file":
		return file.Download
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
