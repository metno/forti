// Package pointdata contains structs and functions for working with
// simpleforecaster's internal data format.
package pointdata

import (
	"time"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/forecast/fortidb/values"
)

type PointData struct {
	Meta *Meta
	Data []values.PointDataCollection
}

// GetData extracts all values for the given parameter
func (p *PointData) GetData(parameter string) map[time.Time]float32 {
	for _, data := range p.Data {
		pm, ok := data.ParameterMeta[parameter]
		if !ok {
			return nil
		}
		ret := make(map[time.Time]float32)
		for i, t := range pm.Times {
			ret[t] = data.Data[pm.SliceFrom+i]
		}
		return ret
	}
	return nil
}

type Meta struct {
	Area       string
	Version    int
	UpdatedAt  time.Time
	NextUpdate time.Time
}

// PointDataCollection contains forecast information for a single point
type PointDataCollection struct {
	ParameterMeta map[string]ParameterMeta
	Data          []float32
}

// ParameterMeta contains metadata about a single parameter
type ParameterMeta struct {
	SliceFrom int         `json:"slice_from"`
	Times     []time.Time `json:"times"`
	Units     string      `json:"units"`
}

type MetaCollection struct {
	Parameters map[string]ParameterMeta `json:"parameters"`
	PointCount int                      `json:"number_of_points"`
}
