package lookup

// #include "geolookup.h"
// #cgo CXXFLAGS: -std=c++11
// #cgo LDFLAGS: -L/usr/local/lib -ls2
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// GeoMap allows looking up nearest index from a list of lat/lon pairs.
type GeoMap struct {
	data unsafe.Pointer
}

// GeoResponse is the response from a Nearest query.
type GeoResponse struct {
	Idx      uint32
	Distance uint32
	Lat      float32
	Long     float32
}

func (gr GeoResponse) String() string {
	return fmt.Sprintf("idx=%d,distance=%d", gr.Idx, gr.Distance)
}

// New creates a new geo map, possibly fetching it from an internal cache.
// Remember to call Free once you are done with the returned object.
func New(latitude, longitude []float32) (*GeoMap, error) {

	if err := sanityCheck(latitude, longitude); err != nil {
		return nil, err
	}
	lat := (*C.float)(&latitude[0])
	lon := (*C.float)(&longitude[0])
	l := &GeoMap{
		data: C.MakePointIndex(lat, lon, C.unsigned(len(latitude))),
	}
	return l, nil
}

func sanityCheck(latitude, longitude []float32) error {
	if len(latitude) != len(longitude) {
		return errors.New("latitude != longitude")
	}
	for _, lat := range latitude {
		if lat < -90 || lat > 90 {
			return fmt.Errorf("latitude is out of range (%f)", lat)
		}
	}
	for _, lon := range longitude {
		if lon < -180 || lon > 180 {
			return fmt.Errorf("longitude is out of range (%f)", lon)
		}
	}
	return nil
}

// Free releases the resources held by this object. Do not call this several times on the same object!
func (l *GeoMap) Free() {
	C.Free(l.data)
}

// Nearest finds the closest index for the given lat/lon.
func (l *GeoMap) Nearest(latitude, longitude float32) (geo GeoResponse, err error) {
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return GeoResponse{}, fmt.Errorf("Invalid value for latitude/longitude (%f/%f)", latitude, longitude)
	}
	gi := C.Nearest(l.data, C.float(latitude), C.float(longitude))
	geo.Idx = uint32(gi.Idx)
	geo.Distance = uint32(gi.Distance)
	geo.Lat = float32(gi.Latitude)
	geo.Long = float32(gi.Longitude)
	return
}
