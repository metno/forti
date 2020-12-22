package forecast

import "github.com/prometheus/client_golang/prometheus"

var groupCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "forti_requested_groups_total",
		Help: "What groups are requested",
	},
	[]string{"group"},
)

func init() {
	prometheus.MustRegister(groupCounter)
}
