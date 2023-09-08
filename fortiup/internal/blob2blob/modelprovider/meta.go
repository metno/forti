package modelprovider

import (
	"time"
)

// Meta contains metadata about a forecast for a single parameter
type Meta struct {
	Parameter string      `json:"parameter"`
	Units     string      `json:"units"`
	Locations uint32      `json:"locations"`
	Times     []time.Time `json:"times"`
}

type CollectedMeta struct {
	Parameters        map[string]ParameterMeta `json:"parameters"`
	NumberOfLocations int                      `json:"number_of_points"`
}

// ParameterMeta contains metadata about a single parameter
type ParameterMeta struct {
	SliceFrom int         `json:"slice_from"`
	Times     []time.Time `json:"times"`
	Units     string      `json:"units"`
}
