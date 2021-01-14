package forecast

import "github.com/prometheus/client_golang/prometheus"

var groupCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "forti_requested_groups_total",
		Help: "What groups are requested",
	},
	[]string{"group"},
)

var distanceHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "forti_distance_to_selected_grid_point",
		Help:    "Distance between requested and selected grid point",
		Buckets: []float64{1_000, 5_000, 10_000, 25_000, 100_000},
	},
	[]string{"group"},
)

var fortiAvailableLatest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "forti_available_latest",
		Help: "Available version for each group (being loaded)",
	},
	[]string{"group"},
)

var fortiAvailableUpdated = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "forti",
		Subsystem: "available",
		Name:      "updated",
		Help:      "Update time of latest version of a dataset",
	},
	[]string{"group"},
)

var fortiActiveLatest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "forti_active_latest",
		Help: "Active version for each group",
	},
	[]string{"group"},
)

var fortiActiveUpdated = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "forti",
		Subsystem: "active",
		Name:      "updated",
		Help:      "Update time of latest version of a dataset",
	},
	[]string{"group"},
)

func init() {
	prometheus.MustRegister(
		groupCounter,
		distanceHistogram,
		fortiActiveLatest,
		fortiActiveUpdated,
		fortiAvailableLatest,
		fortiAvailableUpdated,
	)
}
