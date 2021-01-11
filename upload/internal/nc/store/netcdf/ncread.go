// Package netcdf reads a netcdf file.
package netcdf

// #cgo pkg-config: netcdf
// #include <netcdf.h>
// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// ncMutex syncronizes all access to the netcdf library. This is needed because libnetcdf is not thread safe.
var ncMutex sync.Mutex

// File gives access to a forti netcdf file.
type File C.int

type Float C.float
type Size C.size_t

// Open the the given netcdf file. Not thread safe.
func Open(ncFile string) (File, error) {
	filename := C.CString(ncFile)
	defer C.free(unsafe.Pointer(filename))

	ncMutex.Lock()
	defer ncMutex.Unlock()

	var file C.int
	cerr := C.nc_open(filename, C.NC_NOWRITE, &file)
	if cerr != C.NC_NOERR {
		return 0, mkError(cerr)
	}
	return File(file), nil
}

// Close the file associated with the given netcdf file. Not thread safe.
func (f File) Close() error {
	ncMutex.Lock()
	defer ncMutex.Unlock()

	cerr := C.nc_close(C.int(f))
	if cerr != C.NC_NOERR {
		return mkError(cerr)
	}
	return nil
}

func (f File) GetVariable(name string) (*Variable, error) {
	varname := C.CString(name)
	defer C.free(unsafe.Pointer(varname))

	ncMutex.Lock()
	defer ncMutex.Unlock()

	var varid C.int
	cerr := C.nc_inq_varid(C.int(f), varname, &varid)
	if cerr != C.NC_NOERR {
		return nil, mkError(cerr)
	}

	dims, err := getDimensions(f, varid)
	if err != nil {
		return nil, err
	}

	return &Variable{
		file:       C.int(f),
		varid:      varid,
		Name:       name,
		Dimensions: dims,
	}, nil
}

type Variable struct {
	file  C.int
	varid C.int

	Name       string
	Dimensions []Dimension
}

func mkError(code C.int) error {
	return errors.New(C.GoString(C.nc_strerror(code)))
}

func getDimensions(file File, varid C.int) ([]Dimension, error) {
	var ndims C.int
	if cerr := C.nc_inq_varndims(C.int(file), varid, &ndims); cerr != C.NC_NOERR {
		return nil, mkError(cerr)
	}

	ids := make([]C.int, int(ndims))
	if cerr := C.nc_inq_vardimid(C.int(file), varid, &ids[0]); cerr != C.NC_NOERR {
		return nil, mkError(cerr)
	}

	ret := make([]Dimension, int(ndims))
	for i, id := range ids {
		var size C.size_t
		if cerr := C.nc_inq_dimlen(C.int(file), id, &size); cerr != C.NC_NOERR {
			return nil, mkError(cerr)
		}

		cname := (*C.char)(C.malloc(C.NC_MAX_NAME + 1))
		defer C.free(unsafe.Pointer(cname))
		if cerr := C.nc_inq_dimname(C.int(file), id, cname); cerr != C.NC_NOERR {
			return nil, mkError(cerr)
		}

		ret[i] = Dimension{
			file: file,
			Name: C.GoString(cname),
			Size: int(size),
		}
	}
	return ret, nil
}

func (v *Variable) GetAttributeString(name string) (string, error) {
	attname := C.CString(name)
	defer C.free(unsafe.Pointer(attname))

	ncMutex.Lock()
	defer ncMutex.Unlock()

	var attlen C.size_t
	if cerr := C.nc_inq_attlen(v.file, v.varid, attname, &attlen); cerr != C.NC_NOERR {
		return "", mkError(cerr)
	}

	data := C.malloc(attlen + 1)
	C.bzero(data, attlen+1) // ensure a trailing \0 is set
	defer C.free(data)

	if cerr := C.nc_get_att_text(v.file, v.varid, attname, (*C.char)(data)); cerr != C.NC_NOERR {
		return "", mkError(cerr)
	}

	return C.GoString((*C.char)(data)), nil
}

func (v *Variable) GetFloats(out []Float, start, count []Size) error {
	if len(start) != len(v.Dimensions) || len(count) != len(v.Dimensions) {
		return errors.New("start/count size does not match dimension size")
	}
	outSize := Size(1)
	for _, c := range count {
		outSize *= c
	}
	if int(outSize) != len(out) {
		return fmt.Errorf("invalid size %d, expected %d", len(out), int(outSize))
	}

	ncMutex.Lock()
	defer ncMutex.Unlock()

	if cerr := C.nc_get_vara_float(v.file, v.varid, (*C.size_t)(&start[0]), (*C.size_t)(&count[0]), (*C.float)(&out[0])); cerr != C.NC_NOERR {
		return mkError(cerr)
	}

	return nil
}

func (v *Variable) GetAllFloat() ([]Float, error) {
	totalSize, err := v.TotalDataSize()
	if err != nil {
		return nil, err
	}

	ncMutex.Lock()
	defer ncMutex.Unlock()

	data := make([]Float, totalSize)
	if cerr := C.nc_get_var_float(v.file, v.varid, (*C.float)(&data[0])); cerr != C.NC_NOERR {
		return nil, mkError(cerr)
	}

	return data, nil
}

func (v *Variable) getAllDouble() ([]C.double, error) {
	totalSize, err := v.TotalDataSize()
	if err != nil {
		return nil, err
	}

	data := make([]C.double, totalSize)
	if cerr := C.nc_get_var_double(v.file, v.varid, &data[0]); cerr != C.NC_NOERR {
		return nil, mkError(cerr)
	}

	return data, nil
}

func (v *Variable) GetAllTimes() ([]time.Time, error) {
	ncMutex.Lock()
	defer ncMutex.Unlock()

	values, err := v.getAllDouble()
	if err != nil {
		return nil, err
	}

	ret := make([]time.Time, len(values))
	for i, val := range values {
		ret[i] = time.Unix(int64(val), 0).UTC()
	}

	return ret, nil
}

func (v *Variable) TotalDataSize() (int, error) {
	totalSize := 1
	for _, dim := range v.Dimensions {
		totalSize *= dim.Size
	}
	return totalSize, nil
}

type Dimension struct {
	file File

	Name string
	Size int
}

func (d *Dimension) GetVariable() (*Variable, error) {
	return d.file.GetVariable(d.Name)
}
