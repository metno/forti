package health

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gitlab.met.no/forti/f2/healthz/internal/health/check"
	"gitlab.met.no/forti/f2/healthz/internal/health/config"
)

// ProbeResult is the summed-up result of a set of checks against a number of locations.
type ProbeResult struct {
	Service TypeProbeResult `json:"service"`
	Data    TypeProbeResult `json:"data"`
}

type TypeProbeResult struct {
	OK          bool                `json:"ok"`
	PerformedAt time.Time           `json:"performed_at,omitempty"`
	Locations   map[string][]string `json:"locations,omitempty"`
}

// NewProbeResult creates a new probe result with the given service and data results.
// maxfailedLocations is the maximum number of locations that can fail for the data part of the probe to be considered OK.
// service part will be considered not OK if any location fails.
func NewProbeResult(serviceProblems, dataProblems map[string][]string, maxfailedLocations int) ProbeResult {
	now := time.Now()

	return ProbeResult{
		Service: TypeProbeResult{
			PerformedAt: now,
			OK:          len(serviceProblems) == 0,
			Locations:   serviceProblems,
		},
		Data: TypeProbeResult{
			PerformedAt: now,
			OK:          len(dataProblems) <= maxfailedLocations,
			Locations:   dataProblems,
		},
	}
}

func runProbe(conf *config.ProbeConfiguration) ProbeResult {
	log.Println("Perform probe...")

	serviceProblems := map[string][]string{}
	dataProblems := map[string][]string{}

	for _, request := range conf.GetRequests() {
		log.Printf("Run checks on location %s through url %s\n", request.Name, request.URL)
		service, data := check.Location(conf.Request.Timeout.Duration, request.URL, request.Blueprint)

		if len(service) > 0 {
			serviceProblems[request.Name] = service
			log.Printf("---> Service problems: %v\n", service)
		}
		if len(data) > 0 {
			dataProblems[request.Name] = data
			log.Printf("---> Data problems: %v\n", data)
		}
	}

	probe := NewProbeResult(serviceProblems, dataProblems, conf.Probe.MaxFailedLocations)
	log.Printf("Total result of probe: %v\n", probe)

	return probe
}

func (r ProbeResult) String() string {
	if r.Data.OK && r.Service.OK {
		return "OK"
	}

	msg := ""
	if !r.Service.OK {
		msg += fmt.Sprintf("Service: %s\n", r.Service.String())
	}
	if !r.Data.OK {
		msg += fmt.Sprintf("Data: %s\n", r.Data.String())
	}
	return msg
}

func (tp TypeProbeResult) String() string {
	if tp.OK {
		return "OK"
	}

	messages := make(map[string]int)
	for _, result := range tp.Locations {
		for _, problem := range result {
			messages[problem]++
		}
	}

	uniqueProblems := []string{}
	for p := range messages {
		uniqueProblems = append(uniqueProblems, p)
	}

	return fmt.Sprintf("Not OK, probe with problems for %d locations, caused by these failed checks: %v\n",
		len(tp.Locations), strings.Join(uniqueProblems, ", "))
}
