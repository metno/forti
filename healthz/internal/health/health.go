package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitlab.met.no/forti/f2/healthz/internal/health/config"
)

type Health struct {
	conf *config.CheckConfiguration

	mutex sync.RWMutex

	probeHistory []ProbeResult
	lastProbe    ProbeResult
	isHealthy    bool
}

// New creates a new health instance with the given configuration, with overall health initialized to false.
func New(conf *config.CheckConfiguration) *Health {
	h := &Health{
		conf: conf,
	}

	return h
}

// Start will make health start running regular health checks in the background every minute.
func (h *Health) Start() {
	go func() {
		for {
			h.Probe()
			time.Sleep(time.Minute)
		}
	}()
}

// Health returns the current health, described by the results of the last check and an isHealthy boolean.
func (h *Health) Health() (ProbeResult, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.lastProbe, h.isHealthy
}

// Probe will run a health check immediately.
func (h *Health) Probe() {
	h.setHealth(runProbe(h.conf))
}

// setHealth will decide on the overall health of the system based on the last check and a window of previous checks.
// If last check failed and more than conf.Window.Threshold of the last conf.Window.Size checks
// has failed the result will be not OK.
func (h *Health) setHealth(lastProbe ProbeResult) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.probeHistory = append(h.probeHistory, lastProbe)
	if len(h.probeHistory) > h.conf.ProbeHistory.Size {
		h.probeHistory = h.probeHistory[1:]
	}

	var failed int
	for _, v := range h.probeHistory {
		if !v.OK {
			failed++
		}
	}

	if failed > h.conf.ProbeHistory.MaxFailedProbes &&
		!lastProbe.OK {
		h.isHealthy = false
	} else {
		h.isHealthy = true
	}
	h.lastProbe = lastProbe
}

// ServeSimple will give a plain text description of the last check and returns a 503 if the system is not healthy.
func (h *Health) ServeSimple(w http.ResponseWriter, r *http.Request) {
	lastCheck, isHealthy := h.Health()
	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "text/plain;charset=UTF-8")
	fmt.Fprintln(w, lastCheck)
}

// ServeJSON will give a JSON description of the last check and returns a 503 if the system is not healthy.
func (h *Health) ServeJSON(w http.ResponseWriter, r *http.Request) {
	lastCheck, isHealthy := h.Health()
	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(lastCheck)
}
