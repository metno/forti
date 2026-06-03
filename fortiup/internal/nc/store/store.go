package store

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/metno/forti/fortiup/internal/nc/store/collect"
	"github.com/metno/forti/fortiup/internal/nc/store/netcdf"
	"github.com/metno/forti/fortiup/internal/upload"
	"github.com/metno/forti/fortiup/pkg/fortiblob"
)

func Store(ctx context.Context, u *upload.Uploader, meta *fortiblob.DatasetMeta, files []string) error {
	gridids, err := getGridIds(meta.Area, meta.Version, files)
	if err != nil {
		return err
	}

	for grid, files := range gridids {
		log.Println("store grid ", grid)
		if err := storeGrid(ctx, u, meta.Area, meta.Version, grid, files); err != nil {
			return err
		}
		log.Println("stored ", grid)
	}

	if err := u.SetDatasetMeta(ctx, meta); err != nil {
		return err
	}

	return nil
}

func storeGrid(ctx context.Context, u *upload.Uploader, area string, version int, grid string, files []string) error {
	ncfiles, err := openFiles(files)
	if err != nil {
		return err
	}
	defer func() {
		for _, f := range ncfiles {
			f.Close()
		}
	}()

	var variables []*netcdf.Variable
	for name, file := range ncfiles {
		v, err := file.GetVariable(name)
		if err != nil {
			return fmt.Errorf("unable to netcdf %s: %s", name, err)
		}
		variables = append(variables, v)
	}
	sort.Slice(variables, func(i, j int) bool { return variables[i].Name < variables[j].Name })

	if err := storeLatLon(ctx, u, area, version, grid, ncfiles[variables[0].Name]); err != nil {
		return err
	}

	meta, err := storeData(ctx, u, area, version, grid, variables)
	if err != nil {
		return err
	}

	if err := u.SetGridMeta(ctx, meta, area, version, grid); err != nil {
		return err
	}

	return nil
}

func openFiles(files []string) (map[string]netcdf.File, error) {
	ncfiles := make(map[string]netcdf.File)
	for _, f := range files {
		ncFile, err := netcdf.Open(f)
		if err != nil {
			return nil, fmt.Errorf("unable to open file %s: %s", f, err)
		}
		elements := strings.Split(f, "/")
		variableName := strings.TrimSuffix(elements[len(elements)-1], ".nc")

		ncfiles[variableName] = ncFile
	}
	return ncfiles, nil
}

func storeData(ctx context.Context, u *upload.Uploader, area string, version int, grid string, vars []*netcdf.Variable) (*fortiblob.MetaCollection, error) {
	out, err := u.GetDataStream(ctx, area, version, grid)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriter(out)

	meta, err := collect.Collect(ctx, vars, buf)
	if err != nil {
		return nil, err
	}

	if err := buf.Flush(); err != nil {
		return nil, err
	}

	return meta, out.Close()
}

func storeLatLon(ctx context.Context, u *upload.Uploader, area string, version int, grid string, file netcdf.File) error {
	if err := storeLat(ctx, u, area, version, grid, file); err != nil {
		return err
	}
	if err := storeLon(ctx, u, area, version, grid, file); err != nil {
		return err
	}
	return nil
}

func storeLon(ctx context.Context, u *upload.Uploader, area string, version int, grid string, file netcdf.File) error {
	values, err := getAllFromVariable(file, "lon")
	if err != nil {
		return err
	}
	out, err := u.GetLongitudeStream(ctx, area, version, grid)
	if err != nil {
		return err
	}

	if err := binary.Write(out, binary.LittleEndian, values); err != nil {
		return err
	}

	return out.Close()
}

func storeLat(ctx context.Context, u *upload.Uploader, area string, version int, grid string, file netcdf.File) error {
	values, err := getAllFromVariable(file, "lat")
	if err != nil {
		return err
	}
	out, err := u.GetLatitudeStream(ctx, area, version, grid)
	if err != nil {
		return err
	}

	if err := binary.Write(out, binary.LittleEndian, values); err != nil {
		return err
	}

	return out.Close()
}

func getAllFromVariable(f netcdf.File, variable string) ([]netcdf.Float, error) {
	v, err := f.GetVariable(variable)
	if err != nil {
		return nil, err
	}
	return v.GetAllFloat()
}
