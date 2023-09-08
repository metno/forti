// Package forecast provides the latest forecast for a given location.
package forecast

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.met.no/forti/f2/fortiup/pkg/fortiblob"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/config"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/lookup"
)

// Forecast gives the latest weather forecast for a location.
type Forecast struct {
	store fortiblob.Client

	cfg *config.Configuration

	datasets map[string]*dataset.Dataset
	m        sync.RWMutex
}

// New initializes an object that can be queries for forecasts. It is self-updating.
func New(cfg *config.Configuration) (*Forecast, error) {
	store, err := fortiblob.NewClient(cfg.Source.Bucket)
	if err != nil {
		return nil, err
	}

	f, err := newFromClient(store, cfg)
	if err != nil {
		return nil, err
	}

	go f.run()

	return f, nil
}

func newFromClient(store fortiblob.Client, cfg *config.Configuration) (*Forecast, error) {
	f := &Forecast{
		store:    store,
		cfg:      cfg,
		datasets: make(map[string]*dataset.Dataset),
	}

	if err := f.update(); err != nil {
		return nil, err
	}

	return f, nil
}

// ErrOutsideAllGrids is returned by Forecast.Get if no grids can be found for the given latitude and longitude.
var ErrOutsideAllGrids = errors.New("outside all grids")

// ErrPointTooFarAway is returned by Forecast.Get if a maximum distance to
// returned grid point is defined, and the point has a greater distance
// than that.
var ErrPointTooFarAway = errors.New("returned point is too far away")

// Get returns a forecast for the given latitude and longitude. Returns
// ErrOutsideAllGrids if outside all grids, or ErrPointTooFarAway if grid
// point is too far away.
func (f *Forecast) Get(latitude, longitude float32) (*dataset.LocationData, error) {
	f.m.RLock()
	defer f.m.RUnlock()

	best, err := f.bestArea(latitude, longitude)
	if err != nil {
		return nil, err
	}

	areaCounter.With(prometheus.Labels{"area": best.D.Meta.Area}).Inc()

	locationData, err := best.D.Read(latitude, longitude)
	if err != nil {
		return nil, err
	}
	locationData.GridLocation = best.GridLocation

	return locationData, nil
}

type bestDataset struct {
	D            *dataset.Dataset
	GridLocation dataset.Location
}

func (f *Forecast) bestArea(latitude, longitude float32) (*bestDataset, error) {
	var selected *dataset.Dataset
	var selectedLocation *lookup.GeoResponse
	for _, area := range f.datasets {
		closestLocation, err := area.ClosestGridLocation(latitude, longitude)
		if err != nil {
			return nil, err
		}
		if selectedLocation == nil || closestLocation.Distance < selectedLocation.Distance {
			selected = area
			selectedLocation = closestLocation
		}
	}

	if selected == nil {
		return nil, errors.New("no datasets available")
	}
	if !selected.WithinGeographicArea(latitude, longitude) {
		return nil, ErrOutsideAllGrids
	}
	if !selected.ResponseHasAcceptableDistance(selectedLocation) {
		return nil, ErrPointTooFarAway
	}

	distanceHistogram.With(prometheus.Labels{"area": selected.Meta.Area}).Observe(float64(selectedLocation.Distance))

	return &bestDataset{selected,
		dataset.Location{Lat: selectedLocation.Lat, Long: selectedLocation.Long}}, nil
}

func (f *Forecast) run() {
	for {
		time.Sleep(3 * time.Second)
		if err := f.update(); err != nil {
			// log errors, and retry on next round
			log.Printf("%s - will retry later", err)
		}
	}
}

// update looks up new data from blob store, and loads it.
func (f *Forecast) update() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	latest, err := f.store.Latest(ctx)
	if err != nil {
		return err
	}

	f.m.RLock()
	toAdd := make(map[string]int)
	for _, area := range f.cfg.Areas {
		latestVersion := latest[area]
		current, ok := f.datasets[area]
		if !ok || current.Meta.Version < latestVersion {
			toAdd[area] = latestVersion
		}
	}
	f.m.RUnlock()

	errs := make(chan error)
	for area, version := range toAdd {
		go func(area string, version int) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			errs <- f.load(ctx, area, version)
		}(area, version)
	}
	hasLoadErrors := false
	for area, version := range toAdd {
		if err := <-errs; err != nil {
			log.Printf("unable to load %s/%d: %s", area, version, err)
			hasLoadErrors = true
		}
	}

	if hasLoadErrors {
		return errors.New("unable to load some areas")
	}
	return nil
}

func (f *Forecast) load(ctx context.Context, area string, version int) error {
	log.Printf("available: %s/%d", area, version)

	prometheusLabel := prometheus.Labels{"area": area}

	fortiAvailableLatest.With(prometheusLabel).Set(float64(version))
	fortiAvailableUpdated.With(prometheusLabel).Set(float64(time.Now().Unix()))

	meta, err := f.store.GetMeta(ctx, area, version)
	if err != nil {
		return err
	}

	ds, err := dataset.Download(ctx, f.store, meta, f.cfg)
	if err != nil {
		return err
	}

	f.m.Lock()
	if old, ok := f.datasets[area]; ok {
		old.Close()
	}
	f.datasets[area] = ds
	f.m.Unlock()

	fortiActiveLatest.With(prometheusLabel).Set(float64(version))
	fortiActiveUpdated.With(prometheusLabel).Set(float64(time.Now().Unix()))

	log.Printf("active: %s/%d", area, version)

	return nil
}
