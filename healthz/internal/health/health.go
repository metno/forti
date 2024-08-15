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
	conf *config.ProbeConfiguration

	mutex sync.RWMutex

	probeHistory     []ProbeResult
	lastProbe        ProbeResult
	isDataHealthy    bool
	isServiceHealthy bool
}

// New creates a new health instance with the given configuration, with overall health initialized to false.
func New(conf *config.ProbeConfiguration) *Health {
	h := &Health{
		conf: conf,
	}

	return h
}

// Start will make health start running regular health probes in the background every minute.
func (h *Health) Start() {
	go func() {
		for {
			h.Probe()
			time.Sleep(time.Minute)
		}
	}()
}

// Health returns the current health, described by the results of the last probe and an isHealthy boolean.
func (h *Health) Health() (ProbeResult, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.lastProbe, h.isDataHealthy
}

func (h *Health) Service() (TypeProbeResult, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.lastProbe.Service, h.isServiceHealthy
}

func (h *Health) Data() (TypeProbeResult, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.lastProbe.Service, h.isDataHealthy
}

// Probe will run a health check immediately.
func (h *Health) Probe() {
	h.setHealth(runProbe(h.conf))
}

// setHealth will decide on the overall health of the system based on the last probe and a history of the previous probes.
// If last probe failed and more than conf.probeHistory.MaxFailedProbes of the last conf.probeHistory.Size probes
// has failed, the system will be deemed not healthy.
func (h *Health) setHealth(lastProbe ProbeResult) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.probeHistory = append(h.probeHistory, lastProbe)
	if len(h.probeHistory) > h.conf.ProbeHistory.Size {
		h.probeHistory = h.probeHistory[1:]
	}

	var failed int
	for _, v := range h.probeHistory {
		if !v.Data.OK {
			failed++
		}
	}

	var failedService int
	for _, v := range h.probeHistory {
		if !v.Service.OK {
			failedService++
		}
	}

	if failed > h.conf.ProbeHistory.MaxFailedProbes &&
		!lastProbe.Data.OK {
		h.isDataHealthy = false
	} else {
		h.isDataHealthy = true
	}

	if failedService > h.conf.ProbeHistory.MaxFailedProbes &&
		!lastProbe.Service.OK {
		h.isServiceHealthy = false
	} else {
		h.isServiceHealthy = true
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

// ServeTypeSimple will give a plain text description of the last check and returns a 503 if the system is not healthy.
func (h *Health) ServeTypeSimple(w http.ResponseWriter, r *http.Request) {
	var lastProbe TypeProbeResult
	var isHealthy bool
	if r.PathValue("type") == "service" {
		lastProbe, isHealthy = h.Service()
	} else {
		lastProbe, isHealthy = h.Data()
	}

	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "text/plain;charset=UTF-8")
	fmt.Fprintln(w, lastProbe)
}

// ServeTypeJSON will give a JSON description of the last check and returns a 503 if the system is not healthy.
func (h *Health) ServeTypeJSON(w http.ResponseWriter, r *http.Request) {
	var lastProbe TypeProbeResult
	var isHealthy bool
	if r.PathValue("type") == "service" {
		lastProbe, isHealthy = h.Service()
	} else {
		lastProbe, isHealthy = h.Data()
	}

	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(lastProbe)
}
