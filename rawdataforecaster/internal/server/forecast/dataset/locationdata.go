package dataset

import (
	"time"

	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
)

// LocationData is a struct for holding timeseries data and metadata for a single location from a specific dataset.
type LocationData struct {
	Meta         *Meta
	Data         []values.LocationDataCollection
	GridLocation Location
}

type Location struct {
	Lat  float32
	Long float32
}

// GetData extracts all values for the given parameter
func (p *LocationData) GetData(parameter string) map[time.Time]float32 {
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
