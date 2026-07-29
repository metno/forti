package jsonformat

import (
	"fmt"
	"strings"
)

type GeoJSON struct {
	Type       string    `json:"type"       jsonschema:"enum=Feature"`
	Geometry   Geometry  `json:"geometry"`
	Properties *Forecast `json:"properties,omitempty"`
}

type GeoJSONCoordinate float32
type Geometry struct {
	Type        string              `json:"type"        jsonschema:"enum=Point"`
	Coordinates []GeoJSONCoordinate `json:"coordinates" jsonschema:"minItems=2,maxItems=3,description=Longitude and latitude, with optional altitude as a third element"`
}

func (c GeoJSONCoordinate) MarshalJSON() ([]byte, error) {
	// Max 4 decimals
	s := fmt.Sprintf("%.4f", c)

	// Remove trailing zeros
	s = strings.TrimRight(s, "0")

	// Remove trailing .: E.g: 11. => 11
	s = strings.TrimSuffix(s, ".")

	return []byte(s), nil
}
