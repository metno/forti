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
	conf         *config.ProbeConfiguration
	mutex        sync.RWMutex
	probeHistory []ProbeResult
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

// Probe will run a health check immediately.
func (h *Health) Probe() {
	h.updateProbeHistory(runProbe(h.conf))
}

// setHealth will decide on the overall health of the system based on the last probe and a history of the previous probes.
// If more than conf.probeHistory.MaxFailedProbes of the last conf.probeHistory.Size probes
// has failed, the service part of the system will be deemed not healthy.
// If the data part of the last probe is not OK, the data part of the system will be deemed not healthy.
func (h *Health) updateProbeHistory(lastProbe ProbeResult) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.probeHistory = append(h.probeHistory, lastProbe)
	if len(h.probeHistory) > h.conf.ProbeHistory.Size {
		h.probeHistory = h.probeHistory[1:]
	}
}

type HealthzResponse struct {
	Type      string            `json:"type"`
	IsHealthy bool              `json:"is_healthy"`
	Probes    []TypeProbeResult `json:"probes"`
}

func (hr HealthzResponse) String() string {
	if hr.IsHealthy {
		return "OK"
	}

	msg := "Not OK, due to the probes:\n"
	for _, p := range hr.Probes {
		msg += fmt.Sprintf("---->Performed at %s: ", p.PerformedAt.Format(time.RFC3339))
		msg += fmt.Sprintf("%s\n", p)
	}

	return msg
}

// Health returns the current health, described by the results of the last probe and an isHealthy boolean.
func (h *Health) Health() (ProbeResult, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.probeHistory[len(h.probeHistory)-1], (h.isDataHealthy() && h.isServiceHealthy())
}

func (h *Health) Service() HealthzResponse {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var serviceProbes []TypeProbeResult
	for _, sp := range h.probeHistory {
		serviceProbes = append(serviceProbes, sp.Service)
	}

	return HealthzResponse{
		Type:      "service",
		IsHealthy: h.isServiceHealthy(),
		Probes:    serviceProbes,
	}
}

func (h *Health) Data() HealthzResponse {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var dataProbes []TypeProbeResult
	for _, dp := range h.probeHistory {
		dataProbes = append(dataProbes, dp.Data)
	}

	return HealthzResponse{
		Type:      "data",
		IsHealthy: h.isDataHealthy(),
		Probes:    dataProbes,
	}
}

func (h *Health) isDataHealthy() bool {
	var failedData int
	for _, v := range h.probeHistory {
		if !v.Data.OK {
			failedData++
		}
	}

	return failedData <= h.conf.ProbeHistory.MaxFailedProbes
}

func (h *Health) isServiceHealthy() bool {
	var failedService int
	for _, v := range h.probeHistory {
		if !v.Service.OK {
			failedService++
		}
	}

	return failedService <= h.conf.ProbeHistory.MaxFailedProbes
}

// ServeSimple will give a plain text description of the last probe and returns a 503 if either data or service is not healthy.
func (h *Health) ServeSimple(w http.ResponseWriter, r *http.Request) {
	lastProbe, isHealthy := h.Health()
	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintln(w, lastProbe)
}

// ServeJSON will give a JSON description of the last probe and returns a 503 if either data or service is not healthy.
func (h *Health) ServeJSON(w http.ResponseWriter, r *http.Request) {
	lastProbe, isHealthy := h.Health()

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(lastProbe)
}

// ServeTypeSimple will give a plain text description of either type data or service for all probes in the probe history,
// and returns a 503 if the type is not healthy.
func (h *Health) ServeTypeSimple(w http.ResponseWriter, r *http.Request) {
	var healthzResponse HealthzResponse
	if r.PathValue("type") == "service" {
		healthzResponse = h.Service()
	} else if r.PathValue("type") == "data" {
		healthzResponse = h.Data()
	} else {
		http.Error(w, "Invalid type", http.StatusBadRequest)
	}

	w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	if !healthzResponse.IsHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintln(w, healthzResponse)
}

// ServeTypeJSON will give a json description of either type data or service for all probes in the probe history,
// and returns a 503 if the type is not healthy.
func (h *Health) ServeTypeJSON(w http.ResponseWriter, r *http.Request) {
	var healthzResponse HealthzResponse
	if r.PathValue("type") == "service" {
		healthzResponse = h.Service()
	} else if r.PathValue("type") == "data" {
		healthzResponse = h.Data()
	} else {
		http.Error(w, "Invalid type", http.StatusBadRequest)
	}

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	if !healthzResponse.IsHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(healthzResponse)
}
