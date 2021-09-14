package memory

// #include <stdlib.h>
import "C"
import "unsafe"

// ManuallyAllocatedData contains a manually allocated slice of int16
type manuallyAllocatedData struct {
	Values []int16
	data   unsafe.Pointer
}

// allocate creates a slice of int16, without using go's memory system. The
// memory has to be freed by calling Free().
// Data can be accessed with the Values attribute, but accessing it after Free
// has been called causes undefined behaviour.
func allocate(elements int) *manuallyAllocatedData {
	data := C.calloc(C.ulong(elements), 2)
	values := unsafe.Slice((*int16)(data), elements)

	return &manuallyAllocatedData{
		Values: values,
		data:   data,
	}
}

// Free releases the underlying memory, causing this object to be unusable.
func (d *manuallyAllocatedData) Free() {
	d.Values = nil
	C.free(d.data)
}
