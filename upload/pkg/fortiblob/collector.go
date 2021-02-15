package fortiblob

import "time"

type DatasetMeta struct {
	Area             string          `json:"area"`
	Version          int             `json:"version"`
	TimeUntilNext    time.Duration   `json:"time_until_next"`
	GeographicExtent *GeographicArea `json:"geographic_extent"`
}

// GeographicArea is a specification for of an area on earth. It is specified
// using a Well-Known Text and a Spatial Reference System, expressed as a
// proj4 string.
type GeographicArea struct {
	WKT string `json:"wkt"`
	SRS string `json:"srs"`
}

type MetaCollection struct {
	Parameters    map[string]ParameterMeta `json:"parameters"`
	LocationCount int                      `json:"number_of_points"`
}

// ParameterMeta contains metadata about a forecast for a single parameter
type ParameterMeta struct {
	Units     string      `json:"units"`
	Times     []time.Time `json:"times"`
	SliceFrom int         `json:"slice_from"`
}
