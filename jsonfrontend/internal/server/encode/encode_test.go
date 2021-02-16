package encode

import (
	"encoding/json"
	"testing"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/config"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const configString = `
{
    "cut_forecast": true,
    "location_from_grid": false,
    "http_headers": [
        {
            "key": "Content-Type",
            "value": "application/json; charset=utf-8"
        }
    ],
    "data_expiry_offset": 1800,
    "meta": {
        "#radar_coverage": "precipitation_status"
    },
    "parameters": {
        "instant": {
            "offset": 0,
            "parameters": {
                "air_temperature_2m": "air_temperature"
            }
        }
    }
}
`

func TestEncode(t *testing.T) {
	err := config.InitializeFromString(configString)
	if err != nil {
		t.Errorf("Failed to setup config for encoding; Got error %v", err)
		return
	}

	location := internalprotocol.Location{
		Latitude:  59.124263,
		Longitude: 10.1121323,
	}

	forecast := internalprotocol.Forecast{
		ForecastMeta: &internalprotocol.ForecastMeta{
			UpdatedAt: &timestamppb.Timestamp{},
		},
		ParameterMeta: []*internalprotocol.ParameterMeta{
			{
				Parameter: "air_temperature_2m",
				Units:     "celsius",
				SliceFrom: 0,
				Times:     []*timestamppb.Timestamp{timestamppb.New(time.Now())},
			},
		},
		Data: []float32{12.1},
	}

	geojson, err := Encode(&location, &forecast)
	if err != nil {
		t.Errorf("Expected correct GeoJSON; Got error: %v", err)
	}

	if geojson.Geometry.Coordinates[0] != location.Longitude {
		t.Errorf("Expected latitude coord: %f; Got %f", location.Longitude, geojson.Geometry.Coordinates[0])
	}

	if len(geojson.Properties.Timeseries) == 0 {
		t.Errorf("Got empty timeseries; Expected timeseries of at least length 1")
		return
	}
	airTemp, ok := geojson.Properties.Timeseries[0].Data["instant"].Details["air_temperature"]
	if !ok || airTemp != 12.1 {
		t.Errorf("Expected air_temperature 12.1; Got %f", airTemp)
	}

	// Test serialize geojson
	_, err = json.Marshal(geojson)
	if err != nil {
		t.Errorf("Expected error free json encoding of geojson; Got error: %v.", err)
	}
}
