package fortidb

import (
	"context"
	"log"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "gocloud.dev/blob/fileblob"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
	"gocloud.dev/blob"
)

func Test(t *testing.T) {
	ctx := context.Background()
	_, filename, _, _ := runtime.Caller(0)
	log.Println(filename)
	testPath := filepath.Clean(filepath.Dir(filename) + "/../../../../test/data")

	store, err := blob.OpenBucket(ctx, "file://"+testPath)
	if err != nil {
		t.Fatal(err)
	}

	dataset, err := Download(
		ctx,
		collector.NewClientFromBucket(store),
		&collector.DatasetMeta{
			Area:    "group_b",
			Version: 2,
		},
	)
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
