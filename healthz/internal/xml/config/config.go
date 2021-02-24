package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
)

// Read attempts to reads the given configuration file, and returns a
// matching CheckConfiguration.
func Read(configFile string) (*CheckConfiguration, error) {
	var conf CheckConfiguration
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration file: %s", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&conf); err != nil {
		return nil, fmt.Errorf("unable to read checks config: %s", err)
	}

	return &conf, nil
}

// CheckConfiguration contains a spec for how to execute sanity checks on
// various locationforecast servers- and locations.
type CheckConfiguration struct {
	Headers  map[string]string `json:"headers"`
	Request  Request           `json:"request"`
	Response Response          `json:"response"`
}

type Request struct {
	Protocol     string   `json:"protocol"`
	Servers      []string `json:"servers"`
	PathTemplate string   `json:"path_template"`
}

type Response struct {
	Locations []Location `json:"locations"`
}

type NamedRequest struct {
	Name      string
	URL       *url.URL
	Blueprint Blueprint
}

// GetRequests returns a list of all possible permutations of server address
// and lat/lon.
func (cc *CheckConfiguration) GetRequests() []NamedRequest {
	var ret []NamedRequest

	for _, server := range cc.Request.Servers {
		base := fmt.Sprintf("%s://%s%s", cc.Request.Protocol, server, cc.Request.PathTemplate)
		tmpl, err := template.New("CheckConfiguration").Parse(base)
		if err != nil {
			panic(err)
		}
		for _, loc := range cc.Response.Locations {
			ret = append(ret, loc.getRequest(tmpl))
		}
	}
	return ret
}

// Location is a specification for a location to test.
type Location struct {
	Name      string    `json:"name"`
	Model     string    `json:"model,omitempty"`
	Latitude  float32   `json:"lat"`
	Longitude float32   `json:"lon"`
	Blueprint Blueprint `json:"blueprint"`
}

func (l Location) getRequest(t *template.Template) NamedRequest {
	buffer := bytes.NewBufferString("")
	t.Execute(buffer, l)
	url, err := url.Parse(buffer.String())
	if err != nil {
		panic(err)
	}
	return NamedRequest{
		Name:      l.Name,
		URL:       url,
		Blueprint: l.Blueprint,
	}
}

// Blueprint contains a definition of what kind of data to expect in a
// specific forecast.
type Blueprint struct {
	MaxAge     string                        `json:"max_age"`
	Parameters map[string]CheckSpecification `json:"parameters"`
}

type CheckSpecification struct {
	Attribute    string   `json:"attribute"`
	MinimumCount *uint    `json:"min_count,omitempty"`
	MinimumValue *float32 `json:"min,omitempty"`
	MaximumValue *float32 `json:"max,omitempty"`
	Durations    []int    `json:"durations,omitempty"`
	Optional     bool     `json:"optional,omitempty"`
}
