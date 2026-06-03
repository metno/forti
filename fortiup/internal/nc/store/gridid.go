package store

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"github.com/metno/forti/fortiup/internal/nc/store/netcdf"
)

func getGridIds(area string, version int, files []string) (map[string][]string, error) {
	grids := make(map[string][]string)
	for _, filename := range files {
		f, err := netcdf.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		lat, err := getAllFromVariable(f, "lat")
		if err != nil {
			return nil, err
		}
		lon, err := getAllFromVariable(f, "lon")
		if err != nil {
			return nil, err
		}

		grid := getGridId(lat, lon)

		grids[grid] = append(grids[grid], filename)
	}

	return grids, nil
}

func getGridId(latitudes, longitudes []netcdf.Float) string {
	h := sha256.New()
	if err := binary.Write(h, binary.LittleEndian, latitudes); err != nil {
		panic(err)
	}
	if err := binary.Write(h, binary.LittleEndian, longitudes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
