package check

import (
	"strings"
)

// LocationResult is the result of a set of checks on a single location
type LocationResult struct {
	Data    TypeLocationResult `json:"data"`
	Service TypeLocationResult `json:"service"`
}

type TypeLocationResult struct {
	OK       bool     `json:"ok"`
	Problems []string `json:"problems,omitempty"`
}

func (lr LocationResult) String() string {
	if lr.Data.OK && lr.Service.OK {
		return "OK"
	}

	return "Not OK, caused by: " + strings.Join(append(lr.Data.Problems, lr.Service.Problems...), ", ")
}
