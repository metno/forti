package memory

import (
	"context"
	"log"
	"path/filepath"
	"runtime"
	"testing"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

func Test(t *testing.T) {
	ctx := context.Background()
	_, filename, _, _ := runtime.Caller(0)
	log.Println(filename)
	testPath := filepath.Clean(filepath.Dir(filename) + "/../../../../../../test/data")

	store, err := blob.OpenBucket(ctx, "file://"+testPath)
	if err != nil {
		t.Fatal(err)
	}

	d := newDownloader(collector.NewClientFromBucket(store))

	reader, err := d.Get(ctx,
		&collector.DatasetMeta{
			Area:    "group_a",
			Version: 1,
		},
		"hash_a",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if _, err := reader.Read(4); err == nil {
		t.Error("expected an out-of-bounds error")
	}

	data, err := reader.Read(3)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Data) != 3 {
		t.Errorf("invalid length: %d", len(data.Data))
	}
	if data.Data[0] != 300.0 {
		t.Errorf("unexpected data: %f", data.Data[0])
	}
	if data.Data[1] != 310.0 {
		t.Errorf("unexpected data: %f", data.Data[1])
	}
	if data.Data[2] != 302.0 {
		t.Errorf("unexpected data: %f", data.Data[2])
	}

}
