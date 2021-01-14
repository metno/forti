package collector

import "time"

type DatasetMeta struct {
	Area             string          `json:"area"`
	Version          int             `json:"version"`
	TimeUntilNext    time.Duration   `json:"time_until_next"`
	GeographicExtent *GeographicArea `json:"geographic_extent"`
}

// ParameterMeta contains metadata about a forecast for a single parameter
type ParameterMeta struct {
	Units     string      `json:"units"`
	Times     []time.Time `json:"times"`
	SliceFrom int         `json:"slice_from"`
}

type MetaCollection struct {
	Parameters map[string]ParameterMeta `json:"parameters"`
	PointCount int                      `json:"number_of_points"`
}
