package store

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	"gitlab.met.no/forti/f2/upload/internal/nc/store/collect"
	"gitlab.met.no/forti/f2/upload/internal/nc/store/netcdf"
	"gitlab.met.no/forti/f2/upload/internal/upload"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func Store(ctx context.Context, u *upload.Uploader, meta *fortiblob.DatasetMeta, files []string) error {
	hashes, err := getHashes(meta.Area, meta.Version, files)
	if err != nil {
		return err
	}

	storeResult := make(chan error)
	for hash, files := range hashes {
		go func(hash string, files []string) {
			log.Println("store hash ", hash)
			storeResult <- storeHash(ctx, u, meta.Area, meta.Version, hash, files)
			log.Println("stored ", hash)
		}(hash, files)
	}
	for range hashes {
		if err := <-storeResult; err != nil {
			return err
		}
	}

	if err := u.SetDatasetMeta(ctx, meta); err != nil {
		return err
	}

	return nil
}

func storeHash(ctx context.Context, u *upload.Uploader, area string, version int, hash string, files []string) error {
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

	if err := storeLatLon(ctx, u, area, version, hash, ncfiles[variables[0].Name]); err != nil {
		return err
	}

	meta, err := storeData(ctx, u, area, version, hash, variables)
	if err != nil {
		return err
	}

	if err := u.SetHashMeta(ctx, meta, area, version, hash); err != nil {
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

func storeData(ctx context.Context, u *upload.Uploader, area string, version int, hash string, vars []*netcdf.Variable) (*fortiblob.MetaCollection, error) {
	out, err := u.GetDataStream(ctx, area, version, hash)
	if err != nil {
		return nil, err
	}

	meta, err := collect.Collect(ctx, vars, out)
	if err != nil {
		return nil, err
	}

	return meta, out.Close()
}

func storeLatLon(ctx context.Context, u *upload.Uploader, area string, version int, hash string, file netcdf.File) error {
	if err := storeLat(ctx, u, area, version, hash, file); err != nil {
		return err
	}
	if err := storeLon(ctx, u, area, version, hash, file); err != nil {
		return err
	}
	return nil
}

func storeLon(ctx context.Context, u *upload.Uploader, area string, version int, hash string, file netcdf.File) error {
	values, err := getAllFromVariable(file, "lon")
	if err != nil {
		return err
	}
	out, err := u.GetLongitudeStream(ctx, area, version, hash)
	if err != nil {
		return err
	}

	if err := binary.Write(out, binary.LittleEndian, values); err != nil {
		return err
	}

	return out.Close()
}

func storeLat(ctx context.Context, u *upload.Uploader, area string, version int, hash string, file netcdf.File) error {
	values, err := getAllFromVariable(file, "lat")
	if err != nil {
		return err
	}
	out, err := u.GetLatitudeStream(ctx, area, version, hash)
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
