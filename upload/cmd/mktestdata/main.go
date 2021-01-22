package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func main() {
	workdir := flag.String("workdir", "/tmp", "where to place files")
	area := flag.String("area", "group_a", "group to write")
	version := flag.Int("version", 1, "version to write")
	grid := flag.String("grid", "grid_a", "grid to write")
	latitudes := flag.String("lat", "59,59,60,60", "longitudes")
	longitudes := flag.String("lon", "10,11,10,11", "latitudes")
	parameters := flag.String("parameters", "p1=2,p2=0", "parameters to set, use lik this: p1=<count>,p2=<count>")
	flag.Parse()

	path := fmt.Sprintf("%s/%s/%d/%s/", *workdir, *area, *version, *grid)

	if err := os.MkdirAll(path, os.ModeDir|0770); err != nil {
		log.Fatalln(err)
	}

	meta, err := writeMeta(path, *parameters)
	if err != nil {
		log.Fatalln(err)
	}

	lat, err := extractFloats(*latitudes)
	if err != nil {
		log.Fatalln(err)
	}
	if err := writeBinary(path, "latitude", lat); err != nil {
		log.Fatalln(err)
	}

	lon, err := extractFloats(*longitudes)
	if err != nil {
		log.Fatalln(err)
	}
	if err := writeBinary(path, "longitude", lon); err != nil {
		log.Fatalln(err)
	}

	if len(lat) != len(lon) {
		log.Fatalln("latitudes of differnt length than longitudes")
	}

	if err := writeData(path, meta, lat, lon); err != nil {
		log.Fatalln(err)
	}

	if err := setComplete(*workdir, *area, *version); err != nil {
		log.Fatalln(err)
	}
}

func extractFloats(text string) ([]float32, error) {
	var out []float32
	for _, element := range strings.Split(text, ",") {
		val, err := strconv.ParseFloat(element, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, float32(val))
	}
	return out, nil
}

func writeMeta(path, parameters string) (*fortiblob.MetaCollection, error) {
	meta := fortiblob.MetaCollection{
		Parameters: make(map[string]fortiblob.ParameterMeta),
	}

	re := regexp.MustCompile(`([a-z0-9_]+)=(\d+)`)
	for _, parameter := range strings.Split(parameters, ",") {
		match := re.FindStringSubmatch(parameter)
		if match == nil {
			return nil, fmt.Errorf("invalid parameter spec: %s", parameter)
		}
		count, err := strconv.Atoi(match[2])
		if err != nil {
			return nil, fmt.Errorf("invalid parameter spec: %s", parameter)
		}

		var times []time.Time
		startTime := time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC)
		for i := 0; i < count; i++ {
			times = append(times, startTime.Add(time.Duration(i)*time.Hour))
		}
		pm := fortiblob.ParameterMeta{
			Units:     "u_" + match[1],
			Times:     times,
			SliceFrom: meta.PointCount,
		}
		meta.PointCount += len(times)
		if len(times) == 0 {
			meta.PointCount++
		}
		meta.Parameters[match[1]] = pm
	}
	f, err := os.Create(path + "meta.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(meta); err != nil {
		return nil, err
	}

	return &meta, err
}

func writeData(path string, meta *fortiblob.MetaCollection, lat, lon []float32) error {
	totalSize := len(lat) * meta.PointCount

	data := make([]int16, totalSize)

	for i := range lat {
		for _, pMeta := range meta.Parameters {
			if len(pMeta.Times) == 0 {
				value := (i * 100) + pMeta.SliceFrom
				idx := (i * meta.PointCount) + pMeta.SliceFrom
				data[idx] = int16(value) * 10
			}
			for t := range pMeta.Times {
				value := (i * 100) + (t * 10) + pMeta.SliceFrom
				idx := (i * meta.PointCount) + pMeta.SliceFrom + t
				data[idx] = int16(value) * 10
			}
		}
	}

	return writeBinary(path, "data", data)
}

func writeBinary(path, id string, values interface{}) error {
	f, err := os.Create(path + id)
	if err != nil {
		return err
	}

	if err := binary.Write(f, binary.LittleEndian, values); err != nil {
		return err
	}

	return f.Close()

}

func setComplete(workdir, group string, version int) error {
	path := fmt.Sprintf("%s/%s/%d/complete.json", workdir, group, version)
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	doc := fortiblob.DatasetMeta{
		Area:          group,
		Version:       version,
		TimeUntilNext: time.Hour,
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return setLatest(workdir, group, version)
}

func setLatest(workdir, group string, version int) error {
	path := fmt.Sprintf("%s/latest/%s", workdir, group)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "%d\n", version); err != nil {
		return err
	}
	return f.Close()
}
