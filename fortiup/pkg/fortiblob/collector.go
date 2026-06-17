package fortiblob

import "time"

// DatasetMeta contains metadata about a single area/version combination.
type DatasetMeta struct {

	// Area is the textual identifier for a geographic area.
	Area string `json:"area"`

	// Version is the version number for a forecast for a particular area. A
	// higher version number means a newer forecast.
	Version int `json:"version"`

	// TimeUntilNext is the expected time between when this forecast was ready
	// and when the next one will be available.
	TimeUntilNext time.Duration `json:"time_until_next"`

	// GeographicExtent is an optional description of the geographic area that
	// the locations in this forecast are valid for. If present, it can be
	// used to determine if a request should be served by this area. If not,
	// whatever forecast area has the nearest point can be served.
	GeographicExtent *GeographicArea `json:"geographic_extent"`
}

// GeographicArea is a specification for of an area on earth. It is specified
// using a Well-Known Text and a Spatial Reference System, expressed as a
// proj4 string.
type GeographicArea struct {
	WKT string `json:"wkt"`
	SRS string `json:"srs"`
}

// MetaCollection defines the meaning of the forecast data values. It
// refers to a set of indexable data consisting of int16 values, where index 0
// contains the first piece of data for a particular location.
type MetaCollection struct {

	// Parameters maps parameter names to metadata about that parameter.
	Parameters map[string]ParameterMeta `json:"parameters"`

	// NumberOfPoints is the total number of forecast values for a single
	// location. It is equal to the sum of the length of all times slices
	// under Parameters, and is therefore redundant.
	// We keep it as a separate value to avoid having to calculate it over
	// and over again upon usage.
	NumberOfPoints int `json:"number_of_points"`
}

// ParameterMeta contains metadata about a forecast for a single parameter
type ParameterMeta struct {
	// Units is the units of measure of the data
	Units string `json:"units"`

	// Times are the valid times for each value, in the order in which they
	// appear in the raw data.
	Times []time.Time `json:"times"`

	// SliceFrom describes the offset in the raw data to where these
	// data start.
	SliceFrom int `json:"slice_from"`

	// ScaleFactor describes how the data has been scaled to fit into an int16.
	// Multiply by the ScaleFactor to get the real value from the raw data.
	ScaleFactor float32 `json:"scale_factor,omitempty"`
}
