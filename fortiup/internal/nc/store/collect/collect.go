package collect

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"gitlab.met.no/forti/f2/fortiup/internal/nc/store/netcdf"
	"gitlab.met.no/forti/f2/fortiup/pkg/fortiblob"
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

		scaleFactor := getScaleFactor(variable)
		data := make([]int16, len(floats))
		for i, val := range floats {
			value := math.Round(float64(val) / float64(scaleFactor))
			if value > 32767 || value < -32768 {
				return fmt.Errorf("value for %s is out of range: %f", variable.Name, value)
			}

			data[i] = int16(value)
		}
		if err := binary.Write(out, binary.LittleEndian, data); err != nil {
			return err
		}
	}

	return ctx.Err()
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
