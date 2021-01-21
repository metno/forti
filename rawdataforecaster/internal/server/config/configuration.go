package config

import (
	"context"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/values"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

type DownloadFunction func(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, hash string) (values.Reader, error)

type Configuration struct {
	Bucket                string
	Areas                 []string
	ValueDownloadFunction DownloadFunction
}
