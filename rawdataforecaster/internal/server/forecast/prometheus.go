package forecast

import "github.com/prometheus/client_golang/prometheus"

var areaCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "forti_requested_areas_total",
		Help: "What areas are requested",
	},
	[]string{"area"},
)

var distanceHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "forti_distance_to_selected_grid_point",
		Help:    "Distance between requested and selected grid point",
		Buckets: []float64{1_000, 5_000, 10_000, 25_000, 100_000},
	},
	[]string{"area"},
)

var fortiAvailableLatest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "forti_available_latest",
		Help: "Available version for each area (being loaded)",
	},
	[]string{"area"},
)

var fortiAvailableUpdated = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "forti",
		Subsystem: "available",
		Name:      "updated",
		Help:      "Update time of latest version of a dataset",
	},
	[]string{"area"},
)

var fortiActiveLatest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "forti_active_latest",
		Help: "Active version for each area",
	},
	[]string{"area"},
)

var fortiActiveUpdated = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "forti",
		Subsystem: "active",
		Name:      "updated",
		Help:      "Update time of latest version of a dataset",
	},
	[]string{"area"},
)

func init() {
	prometheus.MustRegister(
		areaCounter,
		distanceHistogram,
		fortiActiveLatest,
		fortiActiveUpdated,
		fortiAvailableLatest,
		fortiAvailableUpdated,
	)
}
