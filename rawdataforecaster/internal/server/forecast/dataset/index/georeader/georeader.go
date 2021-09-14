/*
Package georeader creates lookup.GeoMap objects from a modelprovider connection.
*/
package georeader

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/lookup"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// Reader connects to a model provider, and gets geo data from it
type Reader struct {
	source fortiblob.Client
}

func New(source fortiblob.Client) *Reader {
	return &Reader{source: source}
}

// Get creates a new GeoMap object from the given id
func (r *Reader) Get(ctx context.Context, datasetMeta *fortiblob.DatasetMeta, gridid string) (*lookup.GeoMap, error) {
	latitude, err := r.source.GetLatitude(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}
	defer latitude.Close()
	lats := make([]float32, latitude.Size()/4)
	if err := binary.Read(latitude, binary.LittleEndian, &lats); err != nil {
		return nil, fmt.Errorf("error when getting latitude data from model provider: %s", err)
	}

	longitude, err := r.source.GetLongitude(ctx, datasetMeta, gridid)
	if err != nil {
		return nil, err
	}
	defer longitude.Close()
	lons := make([]float32, longitude.Size()/4)
	if err := binary.Read(longitude, binary.LittleEndian, &lons); err != nil {
		return nil, fmt.Errorf("error when getting longitude data from model provider: %s", err)
	}

	return lookup.New(lats, lons)
}

// Checksum returns a short string that uniquely idenitfies the given set of lat/lon for the given id
func (r *Reader) Checksum(ctx context.Context, datasetMeta *fortiblob.DatasetMeta, gridid string) (string, error) {
	checksumStream := md5.New()

	latitude, err := r.source.GetLatitude(ctx, datasetMeta, gridid)
	if err != nil {
		return "", err
	}
	defer latitude.Close()
	if _, err := io.Copy(checksumStream, latitude); err != nil {
		return "", err
	}

	longitude, err := r.source.GetLongitude(ctx, datasetMeta, gridid)
	if err != nil {
		return "", err
	}
	defer longitude.Close()

	if _, err := io.Copy(checksumStream, longitude); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(checksumStream.Sum(nil)), nil
}
