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

func init() {
	prometheus.MustRegister(groupCounter)
	prometheus.MustRegister(distanceHistogram)
}
