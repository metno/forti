package store

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"gitlab.met.no/forti/f2/upload/internal/nc/store/netcdf"
)

func getHashes(area string, version int, files []string) (map[string][]string, error) {
	hashes := make(map[string][]string)
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

		hash := getHash(lat, lon)

		hashes[hash] = append(hashes[hash], filename)
	}

	return hashes, nil
}

func getHash(latitudes, longitudes []netcdf.Float) string {
	h := sha256.New()
	if err := binary.Write(h, binary.LittleEndian, latitudes); err != nil {
		panic(err)
	}
	if err := binary.Write(h, binary.LittleEndian, longitudes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
