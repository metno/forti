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

	w.Header().Add("Content-Type", "text/plain;charset=UTF-8")
	fmt.Fprintln(w, result)
}

func (c *Checker) ServeJSON(w http.ResponseWriter, r *http.Request) {
	result := c.Check()

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

	log.Println("running checks")
	result := runChecks(c.conf)

	log.Printf("%v\n", result)
	if !result.OK || !c.lastResultOK {
		log.Println(c.lastResult)
	}

	c.lastResultOK = result.OK
	c.lastResult = result
}

func runChecks(conf *config.CheckConfiguration) check.Result {
	result := check.NewResult()

	var failedRequests int
	for _, request := range conf.GetRequests() {

		problems := check.URL(request.URL, request.Blueprint)
		if !problems.OK {
			failedRequests++
			result.Locations[request.Name] = problems
		}
	}
	if failedRequests > conf.Response.MaxFailures {
		result.OK = false
	}

	return result
}
