package health

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gitlab.met.no/forti/f2/healthz/internal/health/json/config"
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
		CheckWindow: config.CheckWindow{
			Size:          1,
			FailThreshold: 0,
		},
		Response: config.Response{
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
			MaxFailures: 0,
		},
	}
	h := New(&conf)
	h.Check()

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
		CheckWindow: config.CheckWindow{
			Size:          3,
			FailThreshold: 1,
		},
		Response: config.Response{
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
			MaxFailures: 0,
		},
	}
	h := New(&conf)

	h.setHealth(runChecks(&conf))
	if !h.isHealthy {
		t.Errorf("Reported failure, but expected ok")
	}

	h.setHealth(runChecks(&conf))
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
