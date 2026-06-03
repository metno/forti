package index

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/metno/forti/fortiup/pkg/fortiblob"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func Test(t *testing.T) {
	bucket, err := blob.OpenBucket(context.TODO(), "mem://test")
	if err != nil {
		t.Fatal(err)
	}
	defer bucket.Close()

	n := addForTest(t, bucket, "area_a", 1, "grid_a", []float32{59, 59, 60, 60}, []float32{11, 10, 11, 10})
	loc, err := n.Nearest(0, 0)
	if err != nil {
		t.Error(err)
	}
	if loc.Idx != 1 {
		t.Errorf("unexpected idx: %d", loc.Idx)
	}

	Free("area_a", 1, "grid_a")

	n2 := addForTest(t, bucket, "area_a", 2, "grid_a", []float32{59, 59, 60, 60}, []float32{11, 10, 11, 10})
	loc, err = n2.Nearest(0, 0)
	if err != nil {
		t.Error(err)
	}
	if loc.Idx != 1 {
		t.Errorf("unexpected idx: %d", loc.Idx)
	}
}

func addForTest(t *testing.T, bucket *blob.Bucket, area string, version int, gridid string, lat, lon []float32) Nearester {
	ctx := context.Background()
	datasetMeta := &fortiblob.DatasetMeta{
		Area:    area,
		Version: version,
	}

	basePath := fmt.Sprintf("%s/%d/%s/", datasetMeta.Area, datasetMeta.Version, gridid)

	addDataToBucket(ctx, bucket, basePath,
		[]float32{59, 59, 60, 60},
		[]float32{11, 10, 11, 10},
	)

	client := fortiblob.NewClientFromBucket(bucket)

	n, err := Add(ctx, client, datasetMeta, gridid)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func addDataToBucket(ctx context.Context, b *blob.Bucket, basePath string, lat, lon []float32) {
	addToBucket(ctx, b, basePath+"latitude", lat)
	addToBucket(ctx, b, basePath+"longitude", lon)
}

func addToBucket(ctx context.Context, b *blob.Bucket, fullPath string, values []float32) {
	w, err := b.NewWriter(ctx, fullPath, nil)
	if err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.LittleEndian, values); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
}
