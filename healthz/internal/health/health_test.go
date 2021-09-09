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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "always fail")
	}))
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Errorf("could not setup test http server; Failed with %v", err)
	}

	conf := config.CheckConfiguration{
		Request: config.Request{
			Protocol: "http",
			Servers: []string{
				serverURL.Host,
			},
			PathTemplate: "/api/forecast/v2/complete?lat={{.Latitude}}&lon={{.Longitude}}",
		},
		Response: config.Response{
			Locations: []config.Location{
				{
					Name:      "AlwaysFail",
					Latitude:  60,
					Longitude: 10,
				},
			},
			MaxFailures: 0,
		},
	}
	checker := NewChecker(&conf)

	req := httptest.NewRequest("GET", server.URL, nil)
	w := httptest.NewRecorder()

	checker.ServeSimple(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503; Got statuscode: %v", resp.StatusCode)
	}
}
