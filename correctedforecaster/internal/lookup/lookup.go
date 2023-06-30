package lookup

// #cgo pkg-config: gdal
// #cgo CXXFLAGS: -std=c++11
// #include "topo.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"fmt"
	"image"
	"sync"
	"unsafe"
)

// Lookup is a gdal-based dataset
type Lookup struct {
	impl       *C.struct_TopoLookup
	mutex      sync.Mutex // gdal library is not thread safe
	projection string
}

// Open tries to use gdal to read the given file
func Open(filename string) (*Lookup, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	impl := C.topo_open(cFilename)
	if impl == nil {
		return nil, errors.New("unable to open dataset")
	}

	return &Lookup{
		impl:       impl,
		projection: C.GoString(C.topo_projection(impl)),
	}, nil
}

// Projection returns the projection information for the given point.
func (l *Lookup) Projection() string {
	return l.projection
}

// Transformation returns a function of converting a latitude/longitude pair
// into x/y coordinates in this object's projection.
func (l *Lookup) Transformation() func(float64, float64) (x, y float64) {
	return func(lat, lon float64) (x, y float64) {
		xy := C.topo_transform(l.impl, C.double(lat), C.double(lon))
		return float64(xy.x), float64(xy.y)
	}
}

// HasDataFor returns true if this data set contains data for the given x/y index.
func (l *Lookup) HasDataFor(x, y float64) bool {
	return C.topo_contains(l.impl, C.double(x), C.double(y)) != 0
}

// Lookup returns the value for the given x/y coordinate.
func (l *Lookup) Lookup(x, y float64) (float32, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var val C.float
	if e := C.topo_lookup(l.impl, &val, C.double(x), C.double(y)); e != C.SUCCESS {
		var err error
		switch e {
		case C.OUT_OF_BOUNDS:
			err = errOutOfBounds
		case C.MISSING_DATA:
			err = errMissingData
		default:
			err = errors.New("unable to read data from file")
		}
		return 0, err
	}

	return float32(val), nil
}

var errOutOfBounds = errors.New("out of bonds")

// IsOutOfBounds returns true if the given error was caused by a Lookup request outside if the grid.
func IsOutOfBounds(err error) bool {
	return err == errOutOfBounds
}

var errMissingData = errors.New("missing data in point")

// IsMissingData returns true if the returned data was within the grid, but contained an explicit missing value.
func IsMissingData(err error) bool {
	return err == errMissingData
}

// GetImage creates a square image centered around the given
// latitude/longitude, with each size of the square in the given size.
func (l *Lookup) GetImage(size int, lat, lon float64) (image.Image, error) {
	img := image.NewGray(image.Rect(0, 0, size, size))

	data := make([]C.float, size*size)
	l.mutex.Lock()
	C.topo_mkgrid(l.impl, &data[0], C.int(size), C.double(lat), C.double(lon))
	l.mutex.Unlock()
	if data == nil {
		return nil, errors.New("unable to create image")
	}

	min := C.float(1000000)
	max := C.float(-10000)
	for _, val := range data {
		if val < min && val != -32767 {
			min = val
		}
		if val > max {
			max = val
		}
	}
	span := max - min
	factor := 256 / span

	fmt.Println(min, max)
	fmt.Println(data[len(data)/2])

	for i, val := range data {
		if val != -32767 {
			img.Pix[i] = uint8((val - min) * factor)
		}
	}

	img.Pix[len(data)/2] = 0

	return img, nil
}

func init() {
	C.topo_init()
}
