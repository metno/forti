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

type Checker struct {
	conf *config.CheckConfiguration

	mutex sync.RWMutex

	lastRun time.Time
	nextRun time.Time

	checkHistory []check.Result
	lastCheck    check.Result
	isHealthy    bool
}

func NewChecker(conf *config.CheckConfiguration) *Checker {
	return &Checker{
		conf: conf,
	}
}

func (c *Checker) ServeSimple(w http.ResponseWriter, r *http.Request) {
	lastCheck, isHealthy := c.Check()
	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "text/plain;charset=UTF-8")
	fmt.Fprintln(w, lastCheck)
}

func (c *Checker) ServeJSON(w http.ResponseWriter, r *http.Request) {
	lastCheck, isHealthy := c.Check()
	if !isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(lastCheck)
}

// Check performs a health check and returns the result and a boolean indicating if the check was successful.
func (c *Checker) Check() (check.Result, bool) {
	now := time.Now()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	shouldRefresh := now.After(c.nextRun)

	if shouldRefresh {
		c.mutex.RUnlock()
		c.refresh()
		c.mutex.RLock()
	}
	return c.lastCheck, c.isHealthy
}

func (c *Checker) refresh() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	// Recheck if this is needed
	if now.Before(c.nextRun) {
		return
	}

	c.lastRun = now
	c.nextRun = now.Add(time.Minute)

	log.Println("Running checks...")
	c.setResult(runChecks(c.conf))
}

// setResult will decide on the overall health of the system based on the last check and a window of previous checks.
// If last check failed and more than conf.Window.Threshold of the last conf.Window.Size checks
// has failed the result will be not OK.
func (c *Checker) setResult(lastCheck check.Result) {
	c.checkHistory = append(c.checkHistory, lastCheck)
	if len(c.checkHistory) > c.conf.CheckWindow.Size {
		c.checkHistory = c.checkHistory[1:]
	}

	var failed int
	for _, v := range c.checkHistory {
		if !v.OK {
			failed++
		}
	}

	if failed > c.conf.CheckWindow.FailThreshold &&
		!lastCheck.OK {
		c.isHealthy = false
	} else {
		c.isHealthy = true
	}
	c.lastCheck = lastCheck
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
