package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(addr, mux)
}

var UpstreamProcessingDuration = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "upstream_processing_duration_seconds",
		Help:      "Time between when we get a request until we are ready to transmit results",
		Buckets:   []float64{0.002, 0.004, 0.008, 0.016, 0.032, 0.064, 0.128, 0.256, 0.512, 1.024},
	},
)

var TotalProcessingDuration = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "total_processing_duration_seconds",
		Help:      "Time between when we get a request until we are done to transmitting results",
		Buckets:   []float64{0.002, 0.004, 0.008, 0.016, 0.032, 0.064, 0.128, 0.256, 0.512, 1.024},
	},
)

var RequestsWithGzip = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "responses_with_gzip",
		Help:      "Response was sent gzipped",
	},
)

var RequestsWithoutGzip = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "responses_without_gzip",
		Help:      "Response was not sent gzipped",
	},
)

var OutsideAllGrids = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "outside_all_grids",
		Help:      "Total number of requests outside coverage area",
	},
)

var PointTooFarAway = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "point_too_far_away",
		Help:      "Total number of requests where the request refers to a point too far away from a valid data point",
	},
)

var UnspecifiedError = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "forti",
		Subsystem: "jsonfrontend",
		Name:      "unspecified_error",
		Help:      "unspecified errors when getting data",
	},
)
