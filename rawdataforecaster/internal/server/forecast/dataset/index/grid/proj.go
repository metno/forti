package grid

// #include "proj_wrapper.h"
// #include <stdlib.h>
// #cgo LDFLAGS: -lproj
import "C"
import (
	"errors"
	"unsafe"
)

// projector converts lat/lon coordinates to coordinates in another projection.
type projector struct {
	pj *C.PJ
}

// newProjector creates a new projector, that converts lat/lon into coordinates in the given coordinate reference system.
func newProjector(crs string) (*projector, error) {
	ctx := C.proj_context_create()
	defer C.proj_context_destroy(ctx)
	def := C.CString(crs)
	defer C.free(unsafe.Pointer(def))
	pj := C.proj_create(ctx, def)
	if pj == nil {
		msg := C.GoString(C.proj_errno_string(C.proj_context_errno(ctx)))
		return nil, errors.New(msg)
	}

	return &projector{
		pj: pj,
	}, nil
}

// Free releases the resources associated with the projector.
func (p *projector) Free() {
	C.proj_destroy(p.pj)
}

// Convert converts the given latitude/longitude to coodinates in another coordinate reference system.
func (p *projector) Convert(ll LatLon) Point {
	coord := C.convert(p.pj, C.double(ll.Longitude), C.double(ll.Latitude))
	return Point{
		X: float64(coord.x),
		Y: float64(coord.y),
	}
}
