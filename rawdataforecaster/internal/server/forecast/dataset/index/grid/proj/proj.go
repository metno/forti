package proj

// #include "proj_wrapper.h"
// #include <stdlib.h>
// #cgo pkg-config: proj
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

var pool map[string][]*C.PJ = make(map[string][]*C.PJ)
var mutex sync.Mutex

type CoordinateTransform struct {
	srs string
	pj  *C.PJ
}

// Get returns a proj coordinate transformation object. It is cached, and
// callers should call Return() on the object when done with it.
// The internal cache is never cleared, since we expect the number of entries
// here to be relatively low.
func Get(srs string) (*CoordinateTransform, error) {
	mutex.Lock()

	inPool := pool[srs]
	size := len(inPool)
	if size == 0 {
		mutex.Unlock()
		return makeCoordinateTransform(srs)
	}

	ret := inPool[size-1]
	pool[srs] = inPool[:size-1]

	mutex.Unlock()

	return &CoordinateTransform{
		srs: srs,
		pj:  ret,
	}, nil
}

func makeCoordinateTransform(srs string) (*CoordinateTransform, error) {
	ctx := C.proj_context_create()
	def := C.CString(srs)
	defer C.free(unsafe.Pointer(def))
	pj := C.proj_create(ctx, def)
	if pj == nil {
		msg := C.GoString(C.proj_errno_string(C.proj_context_errno(ctx)))
		return nil, errors.New(msg)
	}
	return &CoordinateTransform{
		srs: srs,
		pj:  pj,
	}, nil

}

func (ct *CoordinateTransform) Convert(longitude, latitude float64) Point {
	coord := C.convert(ct.pj, C.double(longitude), C.double(latitude))
	return Point{
		X: float64(coord.x),
		Y: float64(coord.y),
	}
}

func (ct *CoordinateTransform) Return() {
	if ct.pj == nil {
		panic("double return on CoordinateTransform")
	}

	mutex.Lock()
	defer mutex.Unlock()

	pool[ct.srs] = append(pool[ct.srs], ct.pj)
	ct.pj = nil
}
