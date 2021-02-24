package check

import (
	"strings"
)

// NewResult initializes a new Result object.
func NewResult() Result {
	return Result{
		OK:        true,
		Locations: make(map[string]LocationResult),
	}
}

// Result is the summed-up result of a set of checks
type Result struct {
	OK        bool                      `json:"ok"`
	Locations map[string]LocationResult `json:"locations,omitempty"`
}

func (r Result) String() string {
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
	for msg := range messages {
		uniqueMessages = append(uniqueMessages, msg)
	}

	return strings.Join(uniqueMessages, ", ")
}

// LocationResult is the result of a set of checks on a single location
type LocationResult struct {
	OK       bool     `json:"ok"`
	Problems []string `json:"problems,omitempty"`
}
