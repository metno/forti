package grid

// #define GEOS_USE_ONLY_R_API
// #include <geos_c.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -lgeos_c
import "C"
import (
	"errors"
	"unsafe"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/grid/proj"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// Area allows performing calculations on a geographic areas.
type Grid struct {
	srs             string
	polygon         *C.GEOSPreparedGeometry
	originalPolygon *C.GEOSGeometry
}

// New creates a new Area struct, with the given Well-Known Text and Coordinate Reference System.
func New(area fortiblob.GeographicArea) (*Grid, error) {

	// Check that the given SRS is valid
	conv, err := proj.Get(area.SRS)
	if err != nil {
		return nil, err
	}
	conv.Return()

	ctx := C.GEOS_init_r()
	defer C.GEOS_finish_r(ctx)
	polygon, err := readGeometry(ctx, area.WKT)
	if err != nil {
		return nil, err
	}

	if typeid := C.GEOSGeomTypeId_r(ctx, polygon); typeid != C.GEOS_POLYGON {
		return nil, errors.New("provided wkt was not a polygon")
	}

	prepared := C.GEOSPrepare_r(ctx, polygon)
	if prepared == nil {
		return nil, errors.New("unable to prepare geometry")
	}

	return &Grid{
		srs:             area.SRS,
		polygon:         prepared,
		originalPolygon: polygon,
	}, nil
}

// Free releases the resources associated with the Area.
func (a *Grid) Free() {
	ctx := C.GEOS_init_r()
	defer C.GEOS_finish_r(ctx)
	C.GEOSPreparedGeom_destroy_r(ctx, a.polygon)
	C.GEOSGeom_destroy_r(ctx, a.originalPolygon)
}

// Contains tells if the given lat/lon is within our polygon. It panics on error.
func (a *Grid) Contains(coord LatLon) bool {
	converter, err := proj.Get(a.srs)
	if err != nil {
		panic(err)
	}
	xy := converter.Convert(coord.Longitude, coord.Latitude)
	converter.Return()

	ctx := C.GEOS_init_r()
	defer C.GEOS_finish_r(ctx)

	point, err := readGeometry(ctx, xy.WKT())
	if err != nil {
		panic(err) // should never happen
	}

	result := byte(C.GEOSPreparedContains_r(ctx, a.polygon, point))
	if result == byte(2) {
		panic("error when checking geometry")
	}
	return result == byte(1)
}

func readGeometry(ctx C.GEOSContextHandle_t, wkt string) (*C.GEOSGeometry, error) {
	reader := C.GEOSWKTReader_create_r(ctx)
	defer C.GEOSWKTReader_destroy_r(ctx, reader)

	cWKT := C.CString(wkt)
	defer C.free(unsafe.Pointer(cWKT))

	geom := C.GEOSWKTReader_read_r(ctx, reader, cWKT)
	if geom == nil {
		return nil, errors.New("invalid geometry")
	}
	return geom, nil
}
