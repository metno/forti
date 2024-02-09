package lookup

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalLookupTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "forti",
			Subsystem: "correctedforecaster",
			Name:      "topolookup_duration_seconds",
			Help:      "The time to lookup an altitude from the topography files.",
			Buckets:   []float64{0.0001, 0.001, 0.01, 0.1, 0.2, 0.5, 1},
		})

	amountOfLookups = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "forti",
			Subsystem: "correctedforecaster",
			Name:      "topolookup_files_amount",
			Help:      "Number of topography file lookups pr request.",
			Buckets:   prometheus.LinearBuckets(float64(1), float64(120), 10),
		},
	)
)

// Collection maintains a prioritized list of Lookup objects, performing
// queries on all added topography files.
type Collection struct {
	lookups         []*Lookup
	transformations map[string]func(float64, float64) (x, y float64)
}

// NewCollection initializes a new, empty Collection.
func NewCollection() *Collection {
	return &Collection{
		transformations: make(map[string]func(float64, float64) (x, y float64)),
	}
}

// Add adds the given files to the collection.
func (c *Collection) Add(filenames ...string) error {
	for _, f := range filenames {
		if err := c.add(f); err != nil {
			return err
		}
	}
	return nil
}

func (c *Collection) add(filename string) error {
	l, err := Open(filename)
	if err != nil {
		return err
	}

	c.lookups = append(c.lookups, l)

	if _, ok := c.transformations[l.Projection()]; !ok {
		c.transformations[l.Projection()] = l.Transformation()
	}

	return nil
}

// Lookup performs a lookup for the given latitude/longitude on all added
// files, returning data for the first match.
// Will return an error if not found. This case can be detected by calling
// IsOutOfBounds on the error.
func (c *Collection) Lookup(latitude, longitude float64) (float32, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		totalLookupTime.Observe(duration)
	}()

	type xy struct {
		x float64
		y float64
	}
	idxs := make(map[string]xy)

	for i, l := range c.lookups {
		idx, ok := idxs[l.Projection()]
		if !ok {
			x, y := l.Transformation()(latitude, longitude)
			idx.x = x
			idx.y = y
			idxs[l.Projection()] = idx
		}

		val, err := l.Lookup(idx.x, idx.y)
		if err != nil {
			if IsOutOfBounds(err) || IsMissingData(err) {
				continue
			}
			amountOfLookups.Observe(float64(i))
			return 0, err
		}
		amountOfLookups.Observe(float64(i))
		return val, nil
	}

	return 0, errOutOfBounds
}
