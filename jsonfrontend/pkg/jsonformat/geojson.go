package jsonformat

import (
	"fmt"
	"strings"
)

type GeoJSON struct {
	Type       string    `json:"type"`
	Geometry   Geometry  `json:"geometry"`
	Properties *Forecast `json:"properties,omitempty"`
}

type GeoJSONCoordinate float32
type Geometry struct {
	Type        string              `json:"type"`
	Coordinates []GeoJSONCoordinate `json:"coordinates"`
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
