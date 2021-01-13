package collector

// GeographicArea is a specification for of an area on earth. It is specified
// using a Well-Known Text and a Spatial Reference System, expressed as a
// proj4 string.
type GeographicArea struct {
	WKT string `json:"wkt"`
	SRS string `json:"srs"`
}
