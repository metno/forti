package encode

import (
	"sort"
	"time"

	"github.com/metno/forti/internalprotocol"
	"github.com/metno/forti/jsonfrontend/internal/server/config"
	"github.com/metno/forti/jsonfrontend/pkg/jsonformat"
	"github.com/metno/forti/internal/radar"
	"github.com/metno/go-weathersymbol"
)

func Encode(forecast *internalprotocol.Forecast) (*jsonformat.GeoJSON, error) {
	ret := baseForecast(forecast.ForecastMeta.GridLocation)

	properties, err := getSerializationForecast(forecast)
	if err != nil {
		return nil, err
	}

	ret.Properties = properties
	if !config.Configuration.SkipAltitude {
		ret.Geometry.Coordinates = append(ret.Geometry.Coordinates, jsonformat.GeoJSONCoordinate(getAltitude(forecast)))
	}

	return ret, nil
}

func EncodeError(forecast *internalprotocol.Forecast, message string) *jsonformat.GeoJSON {
	ret := baseForecast(forecast.ForecastMeta.GridLocation)
	ret.Properties = &jsonformat.Forecast{
		Meta: jsonformat.Metadata{
			UpdatedAt: time.Now().Truncate(time.Minute).UTC(),
			Error:     message,
			Units:     make(map[string]string),
		},
		Timeseries: make([]jsonformat.TimeStep, 0),
	}
	return ret
}

func baseForecast(location *internalprotocol.Location) *jsonformat.GeoJSON {
	return &jsonformat.GeoJSON{
		Type: "Feature",
		Geometry: jsonformat.Geometry{
			Type: "Point",
			Coordinates: []jsonformat.GeoJSONCoordinate{
				jsonformat.GeoJSONCoordinate(location.Longitude),
				jsonformat.GeoJSONCoordinate(location.Latitude),
			},
		},
	}
}

func getSerializationForecast(forecast *internalprotocol.Forecast) (*jsonformat.Forecast, error) {
	return &jsonformat.Forecast{
		Meta:       getMeta(forecast),
		Timeseries: getTimeSteps(forecast),
	}, nil
}

func getAltitude(forecast *internalprotocol.Forecast) float32 {
	for _, p := range forecast.ParameterMeta {
		if len(p.Times) == 1 && p.Times[0].AsTime().Equal(time.Time{}) {
			if p.Parameter == "altitude" {
				return forecast.Data[p.SliceFrom]
			}
		}
	}
	// no altitude in forecast - return 0
	return 0
}

func getMeta(forecast *internalprotocol.Forecast) jsonformat.Metadata {

	myParameters := make(map[string]string)
	for _, timeGroup := range config.Configuration.Parameters {
		for internalParameter, externalParameter := range timeGroup.Parameters {
			myParameters[internalParameter] = externalParameter
		}
	}

	units := make(map[string]string)
	for _, meta := range forecast.ParameterMeta {
		external, ok := myParameters[meta.Parameter]
		if ok {
			units[external] = meta.Units
		}
	}

	return jsonformat.Metadata{
		UpdatedAt:     time.Unix(forecast.ForecastMeta.UpdatedAt.Seconds, 0).UTC(),
		Units:         units,
		RadarCoverage: getRadarCoverage(forecast),
	}
}

func getRadarCoverage(forecast *internalprotocol.Forecast) string {
	if config.Configuration.Meta.RadarCoverage == "" {
		return ""
	}
	for _, meta := range forecast.ParameterMeta {
		if meta.Parameter == config.Configuration.Meta.RadarCoverage {
			c := radar.Coverage(forecast.Data[meta.SliceFrom])
			return c.String()
		}
	}
	return radar.UnknownCoverage.String()
}

func getTimeSteps(forecast *internalprotocol.Forecast) []jsonformat.TimeStep {

	timesteps := make(map[time.Time]map[string]jsonformat.TimestepData)
	for time, values := range GetForecast(forecast).Data {
		for duration, data := range getTimeStep(values) {
			t := time.Add(-duration)
			step, ok := timesteps[t]
			if !ok {
				step = make(map[string]jsonformat.TimestepData)
				timesteps[t] = step
			}
			step[data.Period] = data.Data
		}
	}

	ret := make([]jsonformat.TimeStep, 0, len(timesteps))
	for t, data := range timesteps {
		ret = append(
			ret,
			jsonformat.TimeStep{
				Time: t,
				Data: data,
			},
		)
	}

	sort.Slice(ret,
		func(i, j int) bool {
			return ret[i].Time.Before(ret[j].Time)
		},
	)

	if config.Configuration.CutForecast {
		cutoffTime := time.Now().Truncate(time.Hour)
		for i, data := range ret {
			if !data.Time.Before(cutoffTime) {
				ret = ret[i:]
				break
			}
		}
	}

	return ret
}

type namedTimestep struct {
	Period string
	Data   jsonformat.TimestepData
}

func getTimeStep(values map[string]float32) map[time.Duration]namedTimestep {

	ret := make(map[time.Duration]namedTimestep)

	for timerange, timeGroup := range config.Configuration.Parameters {
		timestepData := jsonformat.TimestepData{
			Summary: getSummary(&timeGroup, values),
			Details: make(jsonformat.ForecastDetails),
		}

		for internalParameter, value := range values {
			if externalParameter, ok := timeGroup.Parameters[internalParameter]; ok {
				timestepData.Details[externalParameter] = jsonformat.SingleDigitFloat(value)
			}
		}

		duration := time.Duration(timeGroup.Offset) * time.Hour

		if timestepData.Summary != nil || len(timestepData.Details) != 0 {
			ret[duration] = namedTimestep{
				Period: timerange,
				Data:   timestepData,
			}
		}
	}

	return ret
}

func getSummary(timeGroup *config.TimeGroup, values map[string]float32) *jsonformat.Summary {
	if timeGroup.Summary != nil {
		hasSummary := false
		var summary jsonformat.Summary

		if value, ok := values[timeGroup.Summary.SymbolCode]; ok {
			symbol := weathersymbol.FromValue(value)
			summary.SymbolCode = symbol.Identifier()
			hasSummary = true
		}
		if value, ok := values[timeGroup.Summary.SymbolConfidence]; ok {
			switch value {
			case 0:
				summary.SymbolConfidence = "certain"
			case 1:
				summary.SymbolConfidence = "somewhat certain"
			case 2:
				summary.SymbolConfidence = "uncertain"
			}
		}
		if hasSummary {
			return &summary
		}
	}
	return nil
}

type Forecast struct {
	Data map[time.Time]map[string]float32 `json:"data"`
}

func GetForecast(forecast *internalprotocol.Forecast) *Forecast {
	ret := Forecast{
		Data: make(map[time.Time]map[string]float32),
	}

	for _, meta := range forecast.ParameterMeta {
		for i, t := range meta.Times {
			time := t.AsTime().UTC()
			timestep, ok := ret.Data[time]
			if !ok {
				timestep = make(map[string]float32)
				ret.Data[time] = timestep
			}
			timestep[meta.Parameter] = forecast.Data[int(meta.SliceFrom)+i]
		}
	}

	return &ret
}
