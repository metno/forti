package forecast

import (
	"context"
	"log"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

func Test(t *testing.T) {
	ctx := context.Background()
	_, filename, _, _ := runtime.Caller(0)
	log.Println(filename)
	testPath := filepath.Clean(filepath.Dir(filename) + "/../../../test/data")

	bucket, err := blob.OpenBucket(ctx, "file://"+testPath)
	if err != nil {
		t.Fatal(err)
	}

	forecast := newFromCollector(collector.NewClientFromBucket(bucket), []string{"group_a", "group_b"})

	pointdata, err := forecast.Get(59, 11)
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
