package upload

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"gitlab.met.no/forti/f2/upload/pkg/collector"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func makeTestingClient() *collector.Client {
	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, "mem://test")
	if err != nil {
		panic(err)
	}
	u := New(bucket)

	area := "group_a"
	version := 1

	dsMeta := collector.DatasetMeta{
		Area:    area,
		Version: version,
	}
	if err := u.SetDatasetMeta(ctx, &dsMeta); err != nil {
		panic(err)
	}

	hash := "hash1"
	paMeta := collector.MetaCollection{
		Parameters: map[string]collector.ParameterMeta{
			"foo": {
				Units: "uFoo",
				Times: []time.Time{
					time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2020, 12, 24, 1, 0, 0, 0, time.UTC),
					time.Date(2020, 12, 24, 2, 0, 0, 0, time.UTC),
				},
				SliceFrom: 0,
			},
			"bar": {
				Units:     "uBar",
				Times:     nil,
				SliceFrom: 3,
			},
		},
		PointCount: 4,
	}
	if err := u.SetHashMeta(ctx, &paMeta, area, version, hash); err != nil {
		panic(err)
	}

	data, err := u.GetDataStream(ctx, area, version, hash)
	if err != nil {
		panic(err)
	}
	dataValues := []int16{
		01, 02, 03, 04,
		11, 12, 13, 14,
		21, 22, 23, 24,
		31, 32, 33, 34,
	}
	if err := binary.Write(data, binary.LittleEndian, dataValues); err != nil {
		panic(err)
	}

	lat, err := u.GetLatitudeStream(ctx, area, version, hash)
	if err != nil {
		panic(err)
	}
	latValues := []float32{
		10, 10, 11, 11,
	}
	if err := binary.Write(lat, binary.LittleEndian, latValues); err != nil {
		panic(err)
	}

	lon, err := u.GetLatitudeStream(ctx, area, version, hash)
	if err != nil {
		panic(err)
	}
	lonValues := []float32{
		59, 60, 59, 60,
	}
	if err := binary.Write(lon, binary.LittleEndian, lonValues); err != nil {
		panic(err)
	}

	return collector.NewClientFromBucket(bucket)
}

func Test(t *testing.T) {
	client := makeTestingClient()
	defer client.Close()

	ctx := context.Background()

	latest, err := client.Latest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	area := "group_a"
	version := latest[area]

	datasetMeta, err := client.GetMeta(ctx, area, version)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetHashes(ctx, datasetMeta)
	if err != nil {
		t.Fatal(err)
	}

}
