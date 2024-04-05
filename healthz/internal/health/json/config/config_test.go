package config

import (
	"encoding/json"
	"testing"
	"time"
)

const nowcastConfig = `{
	"headers": {                                                                                                                                                                                                                     
		 "System": "staging",                                                                                                                                                                                                                                               
		 "Service": "forti",                                                                                                                                                                                                                                                
		 "Responsible": "vegardb,havardf"                                                                                                                                                                                                                                   
	},                                                                                                                                                                                                                                                                     
	"request": {                                                                                                                                                                                                                                                           
	    "protocol": "https",                                                                                                                                                                                                                                               
	    "servers": [                                                                                                                                                                                                                                                       
	        "staging.forti.met.no"                                                                                                                                                                                                                                         
	    ],                                                                                                                                                                                                                                                                 
	    "path_template": "/api/nowcast/v2/complete?lat={{.Latitude}}&lon={{.Longitude}}"                                                                                                                                                                           
	},
	"window": {
		"size": 7,
		"threshold": 3
	},
	"response": {
		"max_failures": 1,
		"locations": [
			{
				"name": "norway nowcast",
				"lat": 63.1,
				"lon": 7.64,
				"blueprint": {
					"max_age": "20m",
					"timeseries": {
						"timeresolution": ["5m"],
						"minduration": "80m"
					},
					"data": {
						"instant": {
							"details": {
								"air_temerature": { "min_count": 1, "min": -110, "max": 110},
								"precipitation_rate": { "min_count": 17},
								"relative_humidity": { "min_count": 1},
								"wind_from_direction": { "min_count": 1},
								"wind_speed": { "min_count": 1},
								"wind_speed_of_gust": { "min_count": 1}
							}
						},	
						"next_1_hours": {
							"summary": {
								"symbol_code": { "min_count": 1 }
							},
							"details": {
								"precipitation_amount": { "min_count": 1 }
							}
						}
					}
				}
			}
		]
	}
}`

func TestDecodeNowcastConfig(t *testing.T) {
	var config CheckConfiguration

	err := json.Unmarshal([]byte(nowcastConfig), &config)
	if err != nil {
		t.Errorf("Expected json decode with no error; Got errors %s", err)
		return
	}

	precipitationRateMinCount := config.Response.Locations[0].Blueprint.Data["instant"].Details["precipitation_rate"].MinimumCount
	if precipitationRateMinCount != 17 {
		t.Errorf("Expected precipitation_rate min count== 17; Got %d", precipitationRateMinCount)
		return
	}

	maxAge := config.Response.Locations[0].Blueprint.MaxAge.Duration
	if maxAge != time.Duration(time.Minute*20) {
		t.Errorf("Expected MaxAge.Duration to be equal to time.Duration(time.Minute * 20); Got this instead: %v", maxAge)
	}
}
