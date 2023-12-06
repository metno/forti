package modelprovider

import "time"

// DataSet contains metadata about a complete dataset (group/version/parameters)
type DataSet struct {
	Group      string   `json:"group"`
	Version    int      `json:"version"`
	Parameters []string `json:"parameters"`

	AvailableAt   time.Time     `json:"available_at,omitempty"`
	TimeUntilNext time.Duration `json:"time-until-next,omitempty"`

	Area GeographicArea `json:"area"`
}

// GeographicArea is a specification for of an area on earth. It is specified
// using a Well-Known Text and a Spatial Reference System, expressed as a
// proj4 string.
type GeographicArea struct {
	WKT string `json:"wkt"`
	SRS string `json:"srs"`
}
