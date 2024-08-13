package health

import (
	"fmt"
	"log"
	"strings"

	"gitlab.met.no/forti/f2/healthz/internal/health/check"
	"gitlab.met.no/forti/f2/healthz/internal/health/config"
)

// NewProbeResult creates a new probe with default values.
func NewProbeResult() ProbeResult {
	return ProbeResult{
		OK:        true,
		Locations: make(map[string]check.LocationResult),
	}
}

// ProbeResult is the summed-up result of a set of checks against a number of locations.
type ProbeResult struct {
	OK        bool                            `json:"ok"`
	Locations map[string]check.LocationResult `json:"locations,omitempty"`
}

func runProbe(conf *config.ProbeConfiguration) ProbeResult {
	log.Println("Perform probe...")

	result := NewProbeResult()

	var failedRequests int
	for _, request := range conf.GetRequests() {
		log.Printf("Run checks on location %s through url %s:\n", request.Name, request.URL)
		locationResult := check.Location(request.URL, request.Blueprint)
		log.Printf("---> Result: %v\n", locationResult)

		if !locationResult.OK {
			failedRequests++
			result.Locations[request.Name] = locationResult
		}
	}
	if failedRequests > conf.Probe.MaxFailedLocations {
		result.OK = false
	}

	log.Printf("Total result of probe: %v\n", result)

	return result
}

func (r ProbeResult) String() string {
	if r.OK {
		return "OK"
	}

	messages := make(map[string]int)
	for _, result := range r.Locations {
		for _, problem := range result.Problems {
			messages[problem]++
		}
	}
	var uniqueMessages []string
	for msg, count := range messages {
		uniqueMessages = append(uniqueMessages, fmt.Sprintf("%s:%d", msg, count))
	}

	return fmt.Sprintf("Not OK,  probe failed for %d locations, caused by these failed checks: %v\n",
		len(r.Locations), strings.Join(uniqueMessages, ", "))
}
