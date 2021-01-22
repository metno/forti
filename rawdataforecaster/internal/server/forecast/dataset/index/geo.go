// Package index handles lookup from a latitude/longitude to an index in a
// downloaded forecast file.
// It uses a global cache to speed up loading.
package index

import (
	"context"
	"fmt"
	"sync"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/georeader"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/lookup"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// Nearester returns the closest index to a given latitude/longitude.
type Nearester interface {
	Nearest(latitude, longitude float32) (index lookup.GeoResponse, err error)
}

type cachedMaps struct {
	geoMap   *lookup.GeoMap
	useCount uint
}

type avg struct {
	Area    string
	Version int
	GridID  string
}

var idCache map[avg]string
var checkSumCache map[string]*cachedMaps
var cacheMutex sync.Mutex

func init() {
	idCache = make(map[avg]string)
	checkSumCache = make(map[string]*cachedMaps)
}

// Add creates or returns a cached lookup object from the given reader and id.
func Add(ctx context.Context, source *fortiblob.Client, datasetMeta *fortiblob.DatasetMeta, gridid string) (Nearester, error) {

	gridReader := georeader.New(source)

	checksum, err := gridReader.Checksum(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, fmt.Errorf("unable to get checksum for %s/%d/%s: %w", datasetMeta.Area, datasetMeta.Version, gridid, err)
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	id := avg{
		Area:    datasetMeta.Area,
		Version: datasetMeta.Version,
		GridID:  gridid,
	}

	if _, ok := idCache[id]; ok {
		free(id)
	}

	cacheEntry, ok := checkSumCache[checksum]
	if !ok {
		return getData(ctx, gridReader, datasetMeta, gridid, checksum)
	}

	idCache[id] = checksum
	cacheEntry.useCount++
	return cacheEntry.geoMap, nil
}

// getData fetches data directly from a model provider, and adds it to the cache
func getData(ctx context.Context, gridReader *georeader.Reader, datasetMeta *fortiblob.DatasetMeta, gridid, checksum string) (Nearester, error) {
	geoMap, err := gridReader.Get(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}

	id := avg{
		Area:    datasetMeta.Area,
		Version: datasetMeta.Version,
		GridID:  gridid,
	}

	idCache[id] = checksum
	checkSumCache[checksum] = &cachedMaps{
		geoMap:   geoMap,
		useCount: 1,
	}
	return geoMap, nil
}

// Free notifies the cache that the given id is no longer in use, and remove
// it from the cache.
// func Free(id gvh) {
// 	cacheMutex.Lock()
// 	defer cacheMutex.Unlock()
// 	free(id)
// }

func free(id avg) {
	checksum, ok := idCache[id]
	if !ok {
		return
	}
	cachedMap, ok := checkSumCache[checksum]
	if !ok {
		// this is a bug
		panic("idCache/checkSumCache mismatch")
	}

	cachedMap.useCount--
	delete(idCache, id)
	if cachedMap.useCount == 0 {
		delete(checkSumCache, checksum)
		cachedMap.geoMap.Free()
	}
}
