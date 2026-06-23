package encode

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/metno/forti/internal/internalprotocol"
	"github.com/metno/forti/jsonfrontend/internal/server/config"
	"github.com/metno/forti/jsonfrontend/pkg/jsonformat"

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

const expectedJSONResponse = `{
	"type": "Feature",
	"geometry": {
		"type": "Point",
		"coordinates": [10.1121, 59.1243, 0]
	},
	"properties": {
		"meta": {
			"updated_at": "1970-01-01T00:00:00Z",
			"units": {
				"air_temperature": "celsius"
			}
		},
		"timeseries": [{
			"time": "0001-01-01T00:00:00Z",
			"data": {
				"instant": {
					"details": {
						"air_temperature": 12.1
					}
				}
			}
		}]
	}
}`

func TestEncode(t *testing.T) {
	err := config.InitializeFromString(configString)
	if err != nil {
		t.Fatalf("Failed to setup config for encoding; Got error %v", err)
	}

	forecast := internalprotocol.Forecast{
		ForecastMeta: &internalprotocol.ForecastMeta{
			UpdatedAt: &timestamppb.Timestamp{},
			GridLocation: &internalprotocol.Location{
				Latitude:  59.124263,
				Longitude: 10.00,
			},
		},
		ParameterMeta: []*internalprotocol.ParameterMeta{
			{
				Parameter: "air_temperature_2m",
				Units:     "celsius",
				SliceFrom: 0,
				Times:     []*timestamppb.Timestamp{timestamppb.New(time.Time{})},
			},
			{
				Parameter: "altitude",
				Units:     "meter",
				SliceFrom: 1,
				Times:     []*timestamppb.Timestamp{timestamppb.New(time.Time{})},
			},
		},
		Data: []float32{12.1, 0},
	}

	geojson, err := Encode(&forecast)
	if err != nil {
		t.Errorf("Expected correct GeoJSON; Got error: %v", err)
	}

	if len(geojson.Properties.Timeseries) == 0 {
		t.Errorf("Got empty timeseries; Expected timeseries of at least length 1")
		return
	}
	airTemp, ok := geojson.Properties.Timeseries[0].Data["instant"].Details["air_temperature"]
	if !ok || airTemp != 12.1 {
		t.Errorf("Expected air_temperature 12.1; Got %f", airTemp)
	}

	payload, err := json.Marshal(geojson)
	if err != nil {
		t.Errorf("Expected error free json encoding of geojson; Got error: %v., payload: %s", err, payload)
	}

	// Decode back and test that coordinates have been trunkated and rounded correctly.
	decoded := jsonformat.GeoJSON{}
	err = json.Unmarshal(payload, &decoded)
	if err != nil {
		t.Errorf("Expected successful decoding of geosjon; Got error: %v", err)
		return
	}

	responseLongitude := jsonformat.GeoJSONCoordinate(10)
	responseLatitude := jsonformat.GeoJSONCoordinate(59.1243)
	if decoded.Geometry.Coordinates[0] != responseLongitude {
		t.Errorf("Expected longitude in decoded response to be %f; Got %f.", responseLongitude, decoded.Geometry.Coordinates[0])
	}

	if decoded.Geometry.Coordinates[1] != responseLatitude {
		t.Errorf("Expected longitude in decoded response to be %f; Got %f.", responseLatitude, decoded.Geometry.Coordinates[1])
	}
}

func TestEncodeSkipAltitude(t *testing.T) {
	err := config.InitializeFromString(configString)
	if err != nil {
		t.Fatalf("Failed to setup config for encoding; Got error %v", err)
	}
	config.Configuration.SkipAltitude = true

	forecast := internalprotocol.Forecast{
		ForecastStatus: internalprotocol.ForecastStatus_OK,
		ForecastMeta: &internalprotocol.ForecastMeta{
			UpdatedAt:  &timestamppb.Timestamp{},
			NextUpdate: &timestamppb.Timestamp{},
			GridLocation: &internalprotocol.Location{
				Latitude:  59,
				Longitude: 11,
			},
		},
		ParameterMeta: nil,
		Data:          nil,
	}

	doc, err := Encode(&forecast)
	if err != nil {
		t.Fatal(err)
	}

	if len(doc.Geometry.Coordinates) != 2 {
		t.Errorf("unexpected geometry: %v", doc.Geometry.Coordinates)
	}
	if doc.Geometry.Coordinates[0] != 11 {
		t.Errorf("unexpected longitude: %f", doc.Geometry.Coordinates[0])
	}
	if doc.Geometry.Coordinates[1] != 59 {
		t.Errorf("unexpected latitude: %f", doc.Geometry.Coordinates[1])
	}
}
