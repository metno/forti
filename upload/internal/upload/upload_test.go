package upload

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func makeTestingClient() *fortiblob.Client {
	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, "mem://test")
	if err != nil {
		panic(err)
	}
	u := New(bucket)

	area := "group_a"
	version := 1

	dsMeta := fortiblob.DatasetMeta{
		Area:    area,
		Version: version,
	}
	if err := u.SetDatasetMeta(ctx, &dsMeta); err != nil {
		panic(err)
	}

	gridid := "gridid1"
	paMeta := fortiblob.MetaCollection{
		Parameters: map[string]fortiblob.ParameterMeta{
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
		LocationCount: 4,
	}
	if err := u.SetGridMeta(ctx, &paMeta, area, version, gridid); err != nil {
		panic(err)
	}

	data, err := u.GetDataStream(ctx, area, version, gridid)
	if err != nil {
		panic(err)
	}
	defer data.Close()
	dataValues := []int16{
		01, 02, 03, 04,
		11, 12, 13, 14,
		21, 22, 23, 24,
		31, 32, 33, 34,
	}
	if err := binary.Write(data, binary.LittleEndian, dataValues); err != nil {
		panic(err)
	}

	lat, err := u.GetLatitudeStream(ctx, area, version, gridid)
	if err != nil {
		panic(err)
	}
	defer lat.Close()
	latValues := []float32{
		10, 10, 11, 11,
	}
	if err := binary.Write(lat, binary.LittleEndian, latValues); err != nil {
		panic(err)
	}

	lon, err := u.GetLatitudeStream(ctx, area, version, gridid)
	if err != nil {
		panic(err)
	}
	defer lon.Close()
	lonValues := []float32{
		59, 60, 59, 60,
	}
	if err := binary.Write(lon, binary.LittleEndian, lonValues); err != nil {
		panic(err)
	}

	return fortiblob.NewClientFromBucket(bucket)
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

	_, err = client.GetGridInfo(ctx, datasetMeta)
	if err != nil {
		t.Fatal(err)
	}

}
