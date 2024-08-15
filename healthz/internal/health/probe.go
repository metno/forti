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
	probeResult := ProbeResult{}
	probeResult.Data.OK = true
	probeResult.Service.OK = true

	return probeResult
}

// ProbeResult is the summed-up result of a set of checks against a number of locations.
type ProbeResult struct {
	Data    TypeProbeResult `json:"data"`
	Service TypeProbeResult `json:"service"`
}

type TypeProbeResult struct {
	OK        bool                                `json:"ok"`
	Locations map[string]check.TypeLocationResult `json:"locations,omitempty"`
}

func runProbe(conf *config.ProbeConfiguration) ProbeResult {
	log.Println("Perform probe...")

	result := NewProbeResult()

	var failedRequests int
	for _, request := range conf.GetRequests() {
		log.Printf("Run checks on location %s through url %s:\n", request.Name, request.URL)
		locationResult := check.Location(request.URL, request.Blueprint)
		log.Printf("---> Result: %v\n", locationResult)

		if !locationResult.Data.OK || !locationResult.Service.OK {
			failedRequests++
		}
		result.Data.Locations[request.Name] = locationResult.Data
		result.Service.Locations[request.Name] = locationResult.Service
	}

	if failedRequests > conf.Probe.MaxFailedLocations {
		result.Data.OK = false
	}

	log.Printf("Total result of probe: %v\n", result)

	return result
}

func (r ProbeResult) String() string {
	if r.Data.OK && r.Service.OK {
		return "OK"
	}

	messages := make(map[string]int)

	if !r.Service.OK {
		for _, result := range r.Service.Locations {
			for _, problem := range result.Problems {
				messages[problem]++
			}
		}

		uniqueProblems := []string{}
		for p := range messages {
			uniqueProblems = append(uniqueProblems, p)
		}
		return fmt.Sprintf("Not OK, probe with service problems for %d locations, caused by these failed checks: %v\n",
			len(r.Service.Locations), strings.Join(uniqueProblems, ", "))

	} else {
		for _, result := range r.Data.Locations {
			for _, problem := range result.Problems {
				messages[problem]++
			}
		}
		uniqueProblems := []string{}
		for p := range messages {
			uniqueProblems = append(uniqueProblems, p)
		}

		return fmt.Sprintf("Not OK, probe with data problems for %d locations, caused by these failed checks: %v\n",
			len(r.Data.Locations), strings.Join(uniqueProblems, ", "))
	}
}
