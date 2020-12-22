package fortiblob

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/pointdata"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func MakeTestingBlob() *blob.Bucket {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "mem://test")
	if err != nil {
		panic(err)
	}
	return bucket
}

func AddToBlob(bucket *blob.Bucket, group string, version int, hash string, parameters map[string]int, latitudes, longitudes []float32) {
	if len(latitudes) != len(longitudes) {
		panic("latitudes and longitudes have different lengths")
	}
	locations := len(latitudes)

	meta := addMeta(bucket, parameters, group, version, hash)
	addData(bucket, meta, locations, group, version, hash)
	addLocations(bucket, group, version, hash, latitudes, longitudes)
}

func addMeta(blob *blob.Bucket, parameters map[string]int, group string, version int, hash string) pointdata.MetaCollection {
	ctx := context.Background()

	meta := pointdata.MetaCollection{
		Parameters: make(map[string]pointdata.ParameterMeta),
	}
	var idx int
	for parameter, size := range parameters {
		startTime := time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC)
		var times []time.Time
		for i := 0; i < size; i++ {
			times = append(times, startTime.Add(time.Duration(i)*time.Hour))
		}
		meta.Parameters[parameter] = pointdata.ParameterMeta{
			SliceFrom: idx,
			Times:     times,
			Units:     fmt.Sprintf("u_%s", parameter),
		}
		idx += size
	}
	meta.PointCount = idx

	w, err := blob.NewWriter(ctx, path(group, version, hash, "meta.json"), nil)
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(w).Encode(&meta); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}

	return meta
}

func addData(blob *blob.Bucket, meta pointdata.MetaCollection, size int, group string, version int, hash string) {
	ctx := context.Background()

	data := make([]int16, meta.PointCount*size)
	for i := 0; i < size; i++ {
		for _, pMeta := range meta.Parameters {
			for j := range pMeta.Times {
				value := (i * 100) + (pMeta.SliceFrom * 10) + j
				idx := (i * meta.PointCount) + pMeta.SliceFrom + j
				data[idx] = int16(value * 10)
			}
		}
	}
	w, err := blob.NewWriter(ctx, path(group, version, hash, "data"), nil)
	if err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.LittleEndian, &data); err != nil {
		panic(err)
	}

	if err := w.Close(); err != nil {
		panic(err)
	}
}

func addLocations(blob *blob.Bucket, group string, version int, hash string, latitudes, longitudes []float32) {
	addRaw(blob, group, version, hash, "latitude", latitudes)
	addRaw(blob, group, version, hash, "longitude", longitudes)
}

func addRaw(blob *blob.Bucket, group string, version int, hash, name string, data []float32) {
	ctx := context.Background()

	w, err := blob.NewWriter(ctx, path(group, version, hash, name), nil)
	if err != nil {
		panic(err)
	}
	if err := binary.Write(w, binary.LittleEndian, &data); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
}
