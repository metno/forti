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

	lastResult   check.Result
	lastResultOK bool
}

func NewChecker(conf *config.CheckConfiguration) *Checker {
	return &Checker{
		conf: conf,
	}
}

func (c *Checker) ServeSimple(w http.ResponseWriter, r *http.Request) {
	result := c.Check()
	if !result.OK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "text/plain;charset=UTF-8")
	fmt.Fprintln(w, result)
}

func (c *Checker) ServeJSON(w http.ResponseWriter, r *http.Request) {
	result := c.Check()
	if !result.OK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

func (c *Checker) Check() check.Result {
	now := time.Now()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	shouldRefresh := now.After(c.nextRun)

	if shouldRefresh {
		c.mutex.RUnlock()
		c.refresh()
		c.mutex.RLock()
	}
	return c.lastResult
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
	result := runChecks(c.conf)

	c.lastResultOK = result.OK
	c.lastResult = result
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
