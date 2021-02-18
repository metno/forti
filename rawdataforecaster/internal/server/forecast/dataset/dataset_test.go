package dataset

import (
	"context"
	"log"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "gocloud.dev/blob/fileblob"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/config"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/lookup"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
	"gocloud.dev/blob"
)

func TestDownloadAndWrite(t *testing.T) {
	dataset, err := downloadDataset()
	if err != nil {
		t.Fatal(err)
	}
	defer dataset.Close()

	pointdata, err := dataset.Read(59, 11)
	if err != nil {
		t.Fatal(err)
	}

	foo := pointdata.GetData("foo")
	if foo == nil {
		t.Fatal("could not find foo")
	}
	if len(foo) != 2 {
		t.Fatalf("invalid size: %d", len(foo))
	}

	if val, ok := foo[time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC)]; val != 100 {
		if !ok {
			t.Error("could not find time")
		} else {
			t.Errorf("unexpected value: %f", val)
		}
	}
	if val, ok := foo[time.Date(2020, 12, 24, 1, 0, 0, 0, time.UTC)]; val != 110 {
		if !ok {
			t.Error("could not find time")
		} else {
			t.Errorf("unexpected value: %f", val)
		}
	}
}

func TestWithinGeographicLimit(t *testing.T) {
	dataset, err := downloadDataset()
	if err != nil {
		t.Fatalf("Expected to get back dataset; Got error: %s", err)
	}

	// Fake user request and lookup georesponse
	reqLat := float32(59.1)
	reqLong := float32(11)

	georesponse := &lookup.GeoResponse{Distance: 500}
	if !dataset.WithinGeographicLimit(georesponse, reqLat, reqLong) {
		t.Error("Expected request to be within maximum distance; Got failed check for geopgraphic limit.")
	}
}

func downloadDataset() (*Dataset, error) {
	ctx := context.Background()
	_, filename, _, _ := runtime.Caller(0)
	log.Println(filename)
	testPath := filepath.Clean(filepath.Dir(filename) + "/../../../../test/data")

	store, err := blob.OpenBucket(ctx, "file://"+testPath)
	if err != nil {
		return nil, err
	}

	return Download(
		ctx,
		fortiblob.NewClientFromBucket(store),
		&fortiblob.DatasetMeta{
			Area:    "group_b",
			Version: 2,
		},
		&config.Configuration{
			Loader: config.Loader{
				Type: "memory",
			},
			MaximumGridPointDistance: 10000,
		},
	)
}
