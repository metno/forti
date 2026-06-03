package collect

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/metno/forti/fortiup/internal/nc/store/netcdf"
	"github.com/metno/forti/fortiup/pkg/fortiblob"
)

func Collect(ctx context.Context, variables []*netcdf.Variable, out io.Writer) (*fortiblob.MetaCollection, error) {
	ret, err := getMetaCollection(ctx, variables)
	if err != nil {
		return nil, err
	}

	if err := collectAllRawData(ctx, out, variables); err != nil {
		return nil, err
	}

	return ret, nil
}

func getMetaCollection(ctx context.Context, variables []*netcdf.Variable) (*fortiblob.MetaCollection, error) {
	ret := fortiblob.MetaCollection{
		Parameters:    make(map[string]fortiblob.ParameterMeta),
		LocationCount: 0,
	}

	for _, v := range variables {
		units, err := v.GetAttributeString("units")
		if err != nil {
			units = ""
		}

		var times []time.Time
		if dim := v.Dimensions[len(v.Dimensions)-1]; dim.Name == "time" {
			timeVar, err := dim.GetVariable()
			if err != nil {
				return nil, fmt.Errorf("unable to get time variable from %s: %w", v.Name, err)
			}
			times, err = timeVar.GetAllTimes()
			if err != nil {
				return nil, fmt.Errorf("unable to extract times from %s: %w", v.Name, err)
			}
		}
		if len(times) == 0 {
			// we still want one time entry if there is no time dimension
			times = append(times, time.Time{})
		}

		meta := fortiblob.ParameterMeta{
			Units:       units,
			Times:       times,
			SliceFrom:   ret.LocationCount,
			ScaleFactor: getScaleFactor(v),
		}

		ret.Parameters[v.Name] = meta
		ret.LocationCount += len(times)
	}

	return &ret, nil
}

func collectAllRawData(ctx context.Context, out io.Writer, variables []*netcdf.Variable) error {
	var caches []*singleParamCache
	for _, v := range variables {
		cache, err := newSingleParamCache(v)
		if err != nil {
			return err
		}
		caches = append(caches, cache)
	}

	for ctx.Err() == nil {
		for i, c := range caches {
			values := c.Next()
			if values == nil {
				return ctx.Err()
			}
			if err := write(variables[i], values, out); err != nil {
				return err
			}
		}
	}

	return ctx.Err()
}

func write(variable *netcdf.Variable, floats []netcdf.Float, out io.Writer) error {
	scaleFactor := getScaleFactor(variable)
	data := make([]int16, len(floats))
	for i, val := range floats {
		value := math.Round(float64(val) / float64(scaleFactor))
		if value > 32767 || value < -32768 {
			return fmt.Errorf("value for %s is out of range: %f", variable.Name, value)
		}

		data[i] = int16(value)
	}
	return binary.Write(out, binary.LittleEndian, data)
}

func getScaleFactor(v *netcdf.Variable) float32 {
	switch v.Name {
	case "sea_floor_depth_below_sea_level":
		return 1
	case "altitude":
		return 1
	default:
		return 0.1
	}
}
