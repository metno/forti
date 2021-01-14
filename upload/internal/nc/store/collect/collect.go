package collect

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"gitlab.met.no/forti/f2/upload/internal/nc/store/netcdf"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func Collect(ctx context.Context, variables []*netcdf.Variable, out io.Writer) (*fortiblob.MetaCollection, error) {
	ret, err := getMetaCollection(ctx, variables)
	if err != nil {
		return nil, err
	}

	size := variables[0].Dimensions[0].Size
	for i := 0; i < size; i++ {
		if err := collectRawData(ctx, out, variables, i); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func getMetaCollection(ctx context.Context, variables []*netcdf.Variable) (*fortiblob.MetaCollection, error) {
	ret := fortiblob.MetaCollection{
		Parameters: make(map[string]fortiblob.ParameterMeta),
		PointCount: 0,
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

		ret.Parameters[v.Name] = fortiblob.ParameterMeta{
			Units:     units,
			Times:     times,
			SliceFrom: ret.PointCount,
		}
		ret.PointCount += len(times)
	}

	return &ret, nil
}

func collectRawData(ctx context.Context, out io.Writer, variables []*netcdf.Variable, idx int) error {
	for _, variable := range variables {
		if len(variable.Dimensions) > 2 {
			return fmt.Errorf("%s has too many dimensions", variable.Name)
		}

		timeSize := 1
		if variable.Dimensions[len(variable.Dimensions)-1].Name == "time" {
			timeSize = variable.Dimensions[len(variable.Dimensions)-1].Size
		}

		floats := make([]netcdf.Float, timeSize)

		start := []netcdf.Size{
			netcdf.Size(idx),
			0,
		}
		count := []netcdf.Size{
			1,
			netcdf.Size(timeSize),
		}
		if err := variable.GetFloats(floats, start[:len(variable.Dimensions)], count[:len(variable.Dimensions)]); err != nil {
			return fmt.Errorf("error when reading %s, idx %d: %w", variable.Name, idx, err)
		}

		data := make([]int16, len(floats))
		for i, val := range floats {
			data[i] = int16(math.Round(float64(val) * 10))
		}
		if err := binary.Write(out, binary.LittleEndian, data); err != nil {
			return err
		}
	}
	return ctx.Err()
}
