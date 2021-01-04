package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.met.no/forti/f2/correctedforecaster/internal/download"
	"gitlab.met.no/forti/f2/correctedforecaster/internal/lookup"
	"gitlab.met.no/forti/f2/jsonfrontend/pkg/jsonformat"
)

func main() {
	workdir := flag.String("workdir", "/data/", "use files in the given directory")
	flag.Parse()

	files, err := download.ListDir(*workdir)
	if err != nil {
		log.Fatalln(err)
	}

	c := lookup.NewCollection()
	if err := c.Add(files...); err != nil {
		log.Fatalln(err)
	}

	log.Println("ready")
	http.Handle("/lookup", &handler{c})
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

type handler struct {
	c *lookup.Collection
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	obj, err := h.getAltitude(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Add("Content-Type", "application/vnd.geo+json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		log.Println(err)
	}
}

func (h *handler) getAltitude(r *http.Request) (*jsonformat.GeoJSON, error) {
	q := r.URL.Query()
	latitude, err := getParam(q, "lat", -90, 90)
	if err != nil {
		return nil, err
	}
	longitude, err := getParam(q, "lon", -180, 180)
	if err != nil {
		return nil, err
	}

	altitude, err := h.c.Lookup(float64(latitude), float64(longitude))
	if err != nil {
		return nil, err
	}

	obj := jsonformat.GeoJSON{
		Type: "Feature",
		Geometry: jsonformat.Geometry{
			Type:        "Point",
			Coordinates: []float32{latitude, longitude, altitude},
		},
	}

	return &obj, nil
}

func getParam(q url.Values, name string, from float32, to float32) (float32, error) {
	val, ok := q[name]
	if !ok || len(val) == 0 {
		return 0, errors.New("Missing value for " + name)
	}
	if len(val) > 1 {
		return 0, errors.New("Too many values for " + name)
	}
	ret, err := strconv.ParseFloat(val[0], 32)
	if err != nil {
		return 0, errors.New("Value for " + name + " cannot be parsed as float")
	}
	value := float32(ret)
	if value < from || value > to {
		return 0, fmt.Errorf("%s must be in range %v to %v", name, from, to)
	}

	return value, nil
}
