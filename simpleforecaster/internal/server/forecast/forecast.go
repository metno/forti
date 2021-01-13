// Package provides the latest forecast for a given location.
package forecast

import (
	"context"
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/forecast/datagroup"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/forecast/datagroup/geo/area"
	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/pointdata"
	"gitlab.met.no/forti/f2/upload/pkg/collector"
)

// Forecast gives the latest weather forecast for a location.
type Forecast struct {
	store  *collector.Client
	groups []string

	datasets map[string]*datagroup.Dataset
	m        sync.RWMutex
}

// New initializes an object that can be queries for forecasts. It is self-updating.
func New(blobURL string, groups []string) (*Forecast, error) {
	store, err := collector.NewClient(blobURL)
	if err != nil {
		return nil, err
	}

	f := newFromCollector(store, groups)

	go f.run()

	return f, nil
}

func newFromCollector(store *collector.Client, groups []string) *Forecast {
	f := &Forecast{
		store:    store,
		groups:   groups,
		datasets: make(map[string]*datagroup.Dataset),
	}

	f.update()

	return f
}

// ErrOutsideAllGrids is returned by Forecast.Get if no grids can be found for the given latitude and longitude.
var ErrOutsideAllGrids = errors.New("outside all grids")

// Get returns a forecast for the given latitude and longitude. Returns ErrOutsideAllGrids if outside all grids.
func (f *Forecast) Get(latitude, longitude float32) (*pointdata.PointData, error) {
	f.m.RLock()
	defer f.m.RUnlock()

	best, err := f.bestGroup(latitude, longitude)
	if err != nil {
		return nil, err
	}

	if best.Area != nil && !best.Area.Contains(area.LatLon{
		Latitude:  float64(latitude),
		Longitude: float64(longitude),
	}) {
		return nil, ErrOutsideAllGrids
	}

	groupCounter.With(prometheus.Labels{"group": best.Meta.Group}).Inc()

	return best.Read(latitude, longitude)
}

func (f *Forecast) bestGroup(latitude, longitude float32) (*datagroup.Dataset, error) {
	selectedDistance := uint(math.MaxUint32)
	var selected *datagroup.Dataset
	for _, datagroup := range f.datasets {
		distance, err := datagroup.DistanceTo(latitude, longitude)
		if err != nil {
			return nil, err
		}
		if distance < selectedDistance {
			selectedDistance = distance
			selected = datagroup
		}
	}
	if selected == nil {
		return nil, errors.New("no datasets available")
	}

	distanceHistogram.With(prometheus.Labels{"group": selected.Meta.Group}).Observe(float64(selectedDistance))

	return selected, nil
}

func (f *Forecast) run() {
	for {
		time.Sleep(30 * time.Second)
		f.update()
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
	for _, group := range f.groups {
		latestVersion := latest[group]
		current, ok := f.datasets[group]
		if !ok || current.Meta.Version < latestVersion {
			toAdd[group] = latestVersion
		}
	}
	f.m.RUnlock()

	var wait sync.WaitGroup
	for group, version := range toAdd {
		wait.Add(1)
		ctx := context.TODO()
		go func(group string, version int) {
			f.load(ctx, group, version)
			wait.Done()
		}(group, version)
	}
	wait.Wait()

	return nil
}

func (f *Forecast) load(ctx context.Context, group string, version int) {
	log.Printf("available: %s/%d", group, version)

	prometheusLabel := prometheus.Labels{"group": group}

	fortiAvailableLatest.With(prometheusLabel).Set(float64(version))
	fortiAvailableUpdated.With(prometheusLabel).Set(float64(time.Now().Unix()))

	meta, err := f.store.GetMeta(ctx, group, version)
	if err != nil {
		log.Fatalln(err)
	}

	datagroup, err := datagroup.Download(ctx, f.store, meta)
	if err != nil {
		log.Fatalln(err)
	}

	f.m.Lock()
	if old, ok := f.datasets[group]; ok {
		old.Close()
	}
	f.datasets[group] = datagroup
	f.m.Unlock()

	fortiActiveLatest.With(prometheusLabel).Set(float64(version))
	fortiActiveUpdated.With(prometheusLabel).Set(float64(time.Now().Unix()))

	log.Printf("active: %s/%d", group, version)
}
