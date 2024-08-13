package check

import (
	"fmt"
	"strings"
)

// NewResult initializes a new Result object.
func NewResult() Probe {
	return Probe{
		OK:        true,
		Locations: make(map[string]LocationResult),
	}
}

// Probe is the summed-up result of a set of checks
type Probe struct {
	OK        bool                      `json:"ok"`
	Locations map[string]LocationResult `json:"locations,omitempty"`
}

func (r Probe) String() string {
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

	return fmt.Sprintf("Not OK,  checks failed for %d locations, caused by these errors: %v\n",
		len(r.Locations), strings.Join(uniqueMessages, ", "))
}

// LocationResult is the result of a set of checks on a single location
type LocationResult struct {
	OK       bool     `json:"ok"`
	Problems []string `json:"problems,omitempty"`
}

func (lr LocationResult) String() string {
	if lr.OK {
		return "OK"
	}
	return "Not OK, caused by: " + strings.Join(lr.Problems, ", ")
}
