// Package forecast provides the latest forecast for a given location.
package forecast

import (
	"context"
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/config"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/dataset/index/grid"
	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/pointdata"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

// Forecast gives the latest weather forecast for a location.
type Forecast struct {
	store *fortiblob.Client
	areas []string

	download config.DownloadFunction

	datasets map[string]*dataset.Dataset
	m        sync.RWMutex
}

// New initializes an object that can be queries for forecasts. It is self-updating.
func New(cfg *config.Configuration) (*Forecast, error) {
	store, err := fortiblob.NewClient(cfg.Bucket)
	if err != nil {
		return nil, err
	}

	f, err := newFromCollector(store, cfg.Areas, cfg.ValueDownloadFunction)
	if err != nil {
		return nil, err
	}

	go f.run()

	return f, nil
}

func newFromCollector(store *fortiblob.Client, areas []string, download config.DownloadFunction) (*Forecast, error) {
	f := &Forecast{
		store:    store,
		areas:    areas,
		download: download,
		datasets: make(map[string]*dataset.Dataset),
	}

	if err := f.update(); err != nil {
		return nil, err
	}

	return f, nil
}

// ErrOutsideAllGrids is returned by Forecast.Get if no grids can be found for the given latitude and longitude.
var ErrOutsideAllGrids = errors.New("outside all grids")

// Get returns a forecast for the given latitude and longitude. Returns ErrOutsideAllGrids if outside all grids.
func (f *Forecast) Get(latitude, longitude float32) (*pointdata.PointData, error) {
	f.m.RLock()
	defer f.m.RUnlock()

	best, err := f.bestArea(latitude, longitude)
	if err != nil {
		return nil, err
	}

	if best.Grid != nil && !best.Grid.Contains(grid.LatLon{
		Latitude:  float64(latitude),
		Longitude: float64(longitude),
	}) {
		return nil, ErrOutsideAllGrids
	}

	areaCounter.With(prometheus.Labels{"area": best.Meta.Area}).Inc()

	return best.Read(latitude, longitude)
}

func (f *Forecast) bestArea(latitude, longitude float32) (*dataset.Dataset, error) {
	selectedDistance := uint(math.MaxUint32)
	var selected *dataset.Dataset
	for _, area := range f.datasets {
		distance, err := area.DistanceTo(latitude, longitude)
		if err != nil {
			return nil, err
		}
		if distance < selectedDistance {
			selectedDistance = distance
			selected = area
		}
	}
	if selected == nil {
		return nil, errors.New("no datasets available")
	}

	distanceHistogram.With(prometheus.Labels{"area": selected.Meta.Area}).Observe(float64(selectedDistance))

	return selected, nil
}

func (f *Forecast) run() {
	for {
		time.Sleep(3 * time.Second)
		if err := f.update(); err != nil {
			// log errors, and retry on next round
			log.Println(err)
		}
	}
}

func (f *Forecast) update() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	latest, err := f.store.Latest(ctx)
	if err != nil {
		return err
	}

	f.m.RLock()
	toAdd := make(map[string]int)
	for _, area := range f.areas {
		latestVersion := latest[area]
		current, ok := f.datasets[area]
		if !ok || current.Meta.Version < latestVersion {
			toAdd[area] = latestVersion
		}
	}
	f.m.RUnlock()

	var wait sync.WaitGroup
	for area, version := range toAdd {
		wait.Add(1)
		ctx := context.TODO()
		go func(area string, version int) {
			f.load(ctx, area, version)
			wait.Done()
		}(area, version)
	}
	wait.Wait()

	return nil
}

func (f *Forecast) load(ctx context.Context, area string, version int) {
	log.Printf("available: %s/%d", area, version)

	prometheusLabel := prometheus.Labels{"area": area}

	fortiAvailableLatest.With(prometheusLabel).Set(float64(version))
	fortiAvailableUpdated.With(prometheusLabel).Set(float64(time.Now().Unix()))

	meta, err := f.store.GetMeta(ctx, area, version)
	if err != nil {
		log.Fatalln(err)
	}

	ds, err := dataset.Download(ctx, f.store, meta, f.download)
	if err != nil {
		log.Fatalln(err)
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
}
