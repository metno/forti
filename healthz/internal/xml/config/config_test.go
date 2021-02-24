package config

import (
	"testing"
)

func TestCreateURL(t *testing.T) {
	cfg := CheckConfiguration{
		Request: Request{
			Protocol:     "https",
			Servers:      []string{"yr.forti.met.no"},
			PathTemplate: "/weatherapi/locationforecast/1.9?lat={{.Latitude}}&lon={{.Longitude}}",
		},
		Response: Response{
			Locations: []Location{
				Location{
					Name:      "somewhere",
					Latitude:  59.0,
					Longitude: 11.01,
				},
			},
		},
	}
	locs := cfg.GetRequests()
	if len(locs) != 1 {
		t.Errorf("got %d urls, expected 1", len(locs))
	}

	if locs[0].Name != "somewhere" {
		t.Errorf("expected name <somewhere>, got <%s>", locs[0].Name)
	}

	expectedURL := "https://yr.forti.met.no/weatherapi/locationforecast/1.9?lat=59&lon=11.01"
	gotURL := locs[0].URL.String()
	if gotURL != expectedURL {
		t.Errorf("got url %s, expected %s", gotURL, expectedURL)
	}
}
