package health

import (
	"fmt"
	"log"
	"strings"

	"gitlab.met.no/forti/f2/healthz/internal/health/check"
	"gitlab.met.no/forti/f2/healthz/internal/health/config"
)

// ProbeResult is the summed-up result of a set of checks against a number of locations.
type ProbeResult struct {
	Data    TypeProbeResult `json:"data"`
	Service TypeProbeResult `json:"service"`
}

type TypeProbeResult struct {
	OK        bool                `json:"ok"`
	Locations map[string][]string `json:"locations,omitempty"`
}

// NewProbeResult creates a new probe result with the given service and data results.
// maxfailedLocations is the maximum number of locations that can fail for the data part of the probe to be considered OK.
// service part will be considered not OK if any location fails.
func NewProbeResult(serviceResults, dataResults map[string][]string, maxfailedLocations int) ProbeResult {
	return ProbeResult{
		Service: TypeProbeResult{
			OK:        len(serviceResults) == 0,
			Locations: serviceResults,
		},
		Data: TypeProbeResult{
			OK:        len(dataResults) >= maxfailedLocations,
			Locations: dataResults,
		},
	}
}

func runProbe(conf *config.ProbeConfiguration) ProbeResult {
	log.Println("Perform probe...")

	serviceResults := map[string][]string{}
	dataResults := map[string][]string{}

	for _, request := range conf.GetRequests() {
		log.Printf("Run checks on location %s through url %s\n", request.Name, request.URL)
		serviceProblems, dataProblems := check.Location(conf.Probe.RequestTimeout.Duration, request.URL, request.Blueprint)

		if len(serviceProblems) > 0 {
			serviceResults[request.Name] = serviceProblems
			log.Printf("---> Service problems: %v\n", serviceProblems)
		}
		if len(dataProblems) > 0 {
			dataResults[request.Name] = dataProblems
			log.Printf("---> Data problems: %v\n", dataProblems)
		}
	}

	probe := NewProbeResult(serviceResults, dataResults, conf.Probe.MaxFailedLocations)
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
