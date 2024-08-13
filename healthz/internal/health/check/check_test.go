package check

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	config "gitlab.met.no/forti/f2/healthz/internal/health/config"
	"gitlab.met.no/forti/f2/jsonfrontend/pkg/jsonformat"
)

func TestJSONCheck(t *testing.T) {
	var forecast jsonformat.GeoJSON
	if err := json.Unmarshal([]byte(faultyJSON), &forecast); err != nil {
		t.Errorf("failed to decode json;expected correct json: Got faulty json: %v", err)
	}

	var count int = 2
	temperatureMin := float32(-90.0)
	temperatureMax := float32(90.0)

	blueprint := config.Blueprint{
		MaxAge: config.Duration{Duration: time.Duration(time.Minute * 60)},
		Timeseries: config.Timeseries{
			Timeresolution: []config.Duration{{Duration: time.Duration(time.Minute * 60)}},
			MinDuration:    config.Duration{Duration: time.Duration(time.Minute * 60 * 48)},
		},
		Data: map[string]config.TimestepData{
			"instant": {
				Details: map[string]config.CheckSpecification{
					"air_temperature":            {MinimumCount: count, MinimumValue: temperatureMin, MaximumValue: temperatureMax},
					"wind_from_direction":        {MinimumCount: count},
					"wind_speed":                 {MinimumCount: count},
					"air_pressure_at_sea_level":  {MinimumCount: count},
					"relative_humidity":          {MinimumCount: count},
					"cloud_area_fraction":        {MinimumCount: count},
					"cloud_area_fraction_low":    {MinimumCount: count},
					"cloud_area_fraction_medium": {MinimumCount: count},
					"cloud_area_fraction_high":   {MinimumCount: count},
					"dew_point_temperature":      {MinimumCount: count},
					"wind_speed_of_gust":         {MinimumCount: count},
					"fog_area_fraction":          {MinimumCount: count},
				},
			},
			"next_1_hours": {
				Details: map[string]config.CheckSpecification{
					"precipitation_amount":         {MinimumCount: count},
					"probability_of_precipitation": {MinimumCount: count},
				},
				Summary: map[string]config.CheckSpecification{
					"symbol_code": {MinimumCount: count},
				},
			},
			"next_6_hours": {
				Details: map[string]config.CheckSpecification{
					"precipitation_amount":         {MinimumCount: count},
					"probability_of_precipitation": {MinimumCount: count},
				},
				Summary: map[string]config.CheckSpecification{
					"symbol_code": {MinimumCount: count},
				},
			},
			"next_12_hours": {
				Details: map[string]config.CheckSpecification{
					"probability_of_precipitation": {MinimumCount: count},
				},
				Summary: map[string]config.CheckSpecification{
					"symbol_code": {MinimumCount: count},
				},
			},
		},
	}

	result := runChecks(forecast.Properties, blueprint)

	if result.OK {
		t.Error("expected errors when checking forecast")
	}

	expectedErrors := 7
	if len(result.Problems) != expectedErrors {
		errString := fmt.Sprintf("Expected %d errors in JSON, got: %d. These are the errors: \n", expectedErrors, len(result.Problems))
		for _, e := range result.Problems {
			errString += fmt.Sprintln(e)
		}
		t.Errorf(errString)
	}
}

const faultyJSON = ` 
{
   "type": "Feature",
   "geometry": {
     "type": "Point",
     "coordinates": [
       11,
       60,
       148
     ]
   },
   "properties": {
     "meta": {
       "updated_at": "2020-12-15T06:51:05Z",
       "units": {
         "air_pressure_at_sea_level": "hPa",
         "air_temperature": "celsius",
         "air_temperature_max": "celsius",
         "air_temperature_min": "celsius",
         "cloud_area_fraction": "%",
         "cloud_area_fraction_high": "%",
         "cloud_area_fraction_low": "%",
         "cloud_area_fraction_medium": "%",
         "dew_point_temperature": "celsius",
         "fog_area_fraction": "%",
         "precipitation_amount": "mm",
         "precipitation_amount_max": "mm",
         "precipitation_amount_min": "mm",
         "probability_of_precipitation": "%",
         "probability_of_thunder": "%",
         "relative_humidity": "%",
         "ultraviolet_index_clear_sky": "1",
         "wind_from_direction": "degrees",
         "wind_speed": "m/s",
         "wind_speed_of_gust": "m/s"
       }
     },
     "timeseries": [
       {
         "time": "2020-12-15T07:00:00Z",
         "data": {
           "instant": {
             "details": {
               "air_pressure_at_sea_level": 1006.1,
               "air_temperature": 91.6,
               "cloud_area_fraction": 99.9,
               "cloud_area_fraction_high": 89.6,
               "cloud_area_fraction_low": 98.7,
               "cloud_area_fraction_medium": 98.4,
               "dew_point_temperature": 1.5,
               "fog_area_fraction": 16.4,
               "relative_humidity": 99.6,
               "ultraviolet_index_clear_sky": 0,
               "wind_from_direction": 28.4,
               "wind_speed": 2.1
             }
           },
           "next_12_hours": {
             "summary": {
               "symbol_code": "rain",
               "symbol_confidence": "certain"
             },
             "details": {
               "probability_of_precipitation": 100
             }
           },
           "next_1_hours": {
             "summary": {
               "symbol_code": "heavyrain"
             },
             "details": {
               "precipitation_amount": 1.1,
               "precipitation_amount_max": 1.2,
               "precipitation_amount_min": 0,
               "probability_of_precipitation": 55.8,
               "probability_of_thunder": 0.1
             }
           },
           "next_6_hours": {
             "summary": {
               "symbol_code": "heavyrain"
             },
             "details": {
               "air_temperature_max": 3.2,
               "air_temperature_min": 2,
               "precipitation_amount": 6.5,
               "precipitation_amount_max": 8.8,
               "precipitation_amount_min": 2.5,
               "probability_of_precipitation": 100
             }
           }
         }
       },
       {
         "time": "2020-12-15T08:10:00Z",
         "data": {
           "instant": {
             "details": {
               "air_pressure_at_sea_level": 1006.3,
               "air_temperature": 2,
               "cloud_area_fraction": 100,
               "cloud_area_fraction_high": 97.8,
               "cloud_area_fraction_low": 99.9,
               "cloud_area_fraction_medium": 96.7,
               "dew_point_temperature": 1.7,
               "fog_area_fraction": 5.1,
               "relative_humidity": 98.9,
               "ultraviolet_index_clear_sky": 0,
               "wind_from_direction": 51.2,
               "wind_speed": 2,
               "wind_speed_of_gust": 3
             }
           },
           "next_12_hours": {
             "summary": {
               "symbol_code": "rain",
               "symbol_confidence": "certain"
             },
             "details": {
               "probability_of_precipitation": 100
             }
           },
           "next_1_hours": {
             "details": {
               "precipitation_amount": 0.5,
               "precipitation_amount_max": 1,
               "precipitation_amount_min": 0,
               "probability_of_precipitation": 71.6,
               "probability_of_thunder": 0.1
             }
           },
           "next_6_hours": {
             "summary": {
               "symbol_code": "heavyrain"
             },
             "details": {
               "air_temperature_max": 3.5,
               "air_temperature_min": 2.1,
               "precipitation_amount": 6.7,
               "precipitation_amount_max": 9.5,
               "precipitation_amount_min": 2.9,
               "probability_of_precipitation": 100
             }
           }
         }
       }
    ]
  }
}
`
