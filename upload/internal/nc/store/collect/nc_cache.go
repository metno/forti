package collect

import (
	"fmt"

	"gitlab.met.no/forti/f2/upload/internal/nc/store/netcdf"
)

type singleValueCache struct {
	data      []netcdf.Float
	chunkSize int
	idx       int
}

func newSingleValueCache(v *netcdf.Variable) (*singleValueCache, error) {
	if len(v.Dimensions) > 2 {
		return nil, fmt.Errorf("%s has too many dimensions", v.Name)
	}

	timeSize := 1
	if v.Dimensions[len(v.Dimensions)-1].Name == "time" {
		timeSize = v.Dimensions[len(v.Dimensions)-1].Size
	}

	data, err := v.GetAllFloat()
	if err != nil {
		return nil, err
	}

	return &singleValueCache{
		data:      data,
		chunkSize: timeSize,
	}, nil
}

// Next returns the next chunk, or nil if there are no more chunks.
func (c *singleValueCache) Next() []netcdf.Float {
	if c.idx >= len(c.data) {
		return nil
	}

	ret := c.data[c.idx : c.idx+c.chunkSize]
	c.idx += c.chunkSize
	return ret
}
