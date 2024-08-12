package health

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"gitlab.met.no/forti/f2/healthz/internal/health/json/check"
	"gitlab.met.no/forti/f2/healthz/internal/health/json/config"
)

type Health struct {
	conf *config.CheckConfiguration

	mutex sync.RWMutex

	checkHistory []check.Result
	lastCheck    check.Result
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
			h.Check()
			time.Sleep(time.Minute)
		}
	}()
}

// Health returns the current health, described by the results of the last check and an isHealthy boolean.
func (h *Health) Health() (check.Result, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.lastCheck, h.isHealthy
}

// Check will run a health check immediately.
func (h *Health) Check() {
	log.Println("Running checks...")
	h.setHealth(runChecks(h.conf))
}

// setHealth will decide on the overall health of the system based on the last check and a window of previous checks.
// If last check failed and more than conf.Window.Threshold of the last conf.Window.Size checks
// has failed the result will be not OK.
func (h *Health) setHealth(lastCheck check.Result) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.checkHistory = append(h.checkHistory, lastCheck)
	if len(h.checkHistory) > h.conf.CheckWindow.Size {
		h.checkHistory = h.checkHistory[1:]
	}

	var failed int
	for _, v := range h.checkHistory {
		if !v.OK {
			failed++
		}
	}

	if failed > h.conf.CheckWindow.FailThreshold &&
		!lastCheck.OK {
		h.isHealthy = false
	} else {
		h.isHealthy = true
	}
	h.lastCheck = lastCheck
}

func runChecks(conf *config.CheckConfiguration) check.Result {
	result := check.NewResult()

	var failedRequests int
	for _, request := range conf.GetRequests() {
		log.Printf("Check %s on url %s ...\n", request.Name, request.URL)
		locationResult := check.URL(request.URL, request.Blueprint)
		log.Printf("---> Result: %v\n", locationResult)

		if !locationResult.OK {
			failedRequests++
			result.Locations[request.Name] = locationResult
		}
	}
	if failedRequests > conf.Response.MaxFailures {
		result.OK = false
	}

	log.Printf("Total result of all checks: %v\n", result)

	return result
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
