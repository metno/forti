package health

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gitlab.met.no/forti/f2/healthz/internal/health/config"
)

func TestForecastServiceUnavailable(t *testing.T) {
	serverURL, err := MockServerURL()
	if err != nil {
		t.Errorf("failed to get mock server: %s", err)
	}

	conf := config.CheckConfiguration{
		Request: config.Request{
			Protocol: "http",
			Servers: []string{
				serverURL.Host,
			},
			PathTemplate: "/api/forecast/v2/complete?lat={{.Latitude}}&lon={{.Longitude}}",
		},
		ProbeHistory: config.ProbeHistory{
			Size:            1,
			MaxFailedProbes: 0,
		},
		Probe: config.Probe{
			Locations: []config.Location{
				{
					Name:      "AlwaysFail",
					Latitude:  60,
					Longitude: 10,
				},
				{
					Name:      "AlwaysFail2",
					Latitude:  50,
					Longitude: 10,
				},
			},
			MaxFailedLocations: 0,
		},
	}
	h := New(&conf)
	h.Probe()

	req := httptest.NewRequest("GET", serverURL.RequestURI(), nil)
	w := httptest.NewRecorder()

	h.ServeSimple(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503; Got statuscode: %v", resp.StatusCode)
	}
}

func TestCheckWithFailureWindow(t *testing.T) {
	serverURL, err := MockServerURL()
	if err != nil {
		t.Errorf("failed to get mock server: %s", err)
	}

	conf := config.CheckConfiguration{
		Request: config.Request{
			Protocol: "http",
			Servers: []string{
				serverURL.Host,
			},
			PathTemplate: "/api/forecast/v2/complete?lat={{.Latitude}}&lon={{.Longitude}}",
		},
		ProbeHistory: config.ProbeHistory{
			Size:            3,
			MaxFailedProbes: 1,
		},
		Probe: config.Probe{
			Locations: []config.Location{
				{
					Name:      "AlwaysFail",
					Latitude:  60,
					Longitude: 10,
				},
				{
					Name:      "AlwaysFail2",
					Latitude:  50,
					Longitude: 10,
				},
			},
			MaxFailedLocations: 0,
		},
	}
	h := New(&conf)

	h.setHealth(runProbe(&conf))
	if !h.isHealthy {
		t.Errorf("Reported failure, but expected ok")
	}

	h.setHealth(runProbe(&conf))
	if h.isHealthy {
		t.Errorf("Reported ok, but expected failure.")
	}
}

func MockServerURL() (*url.URL, error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "always fail")
	}))
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		return nil, fmt.Errorf("could not setup test http server; Failed with %v", err)
	}

	return serverURL, nil
}
