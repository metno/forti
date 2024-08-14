package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"regexp"
	"time"
)

// Read attempts to reads the given configuration file, and returns a
// matching ProbeConfiguration.
func Read(configFile string) (*ProbeConfiguration, error) {
	var conf ProbeConfiguration
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration file: %s", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&conf); err != nil {
		return nil, fmt.Errorf("unable to read checks config: %s", err)
	}
	setDefaultProbeHistory(&conf)

	return &conf, nil
}

// setDefaultProbeHistory sets default probe history if size is zero or negative.
// default probe history will accept 1 failure in 10 probes.
func setDefaultProbeHistory(conf *ProbeConfiguration) {
	if conf.ProbeHistory.Size < 1 {
		conf.ProbeHistory.Size = 10
		conf.ProbeHistory.MaxFailedProbes = 1
	}
}

// ProbeConfiguration contains a spec for how to execute sanity checks on
// various locationforecast servers- and locations.
type ProbeConfiguration struct {
	Headers      map[string]string `json:"headers"`
	ProbeHistory ProbeHistory      `json:"probe_history"`
	Request      Request           `json:"request"`
	Probe        Probe             `json:"probe"`
}

// Request is a specification for how to construct a Forti request.
type Request struct {
	Protocol     string   `json:"protocol"`
	Servers      []string `json:"servers"`
	PathTemplate string   `json:"path_template"`
}

type ProbeHistory struct {
	Size            int `json:"size"`              // Size is the number of the most recent probes that are kept in memory.
	MaxFailedProbes int `json:"max_failed_probes"` // If more than MaxFailedProbes have failed, the system is considered unhealthy.
}

type Probe struct {
	MaxFailedLocations int        `json:"max_failed_locations"` // MaxFailedLocations is the number of locations that can fail before the probe is considered failed.
	Locations          []Location `json:"locations"`            // Locations are the check specifications for a list of locations.
}

// Problems returns a list of errors in the configuration, but only errors
// that are so severe that checks cannot be based on this config.
func (cc *ProbeConfiguration) Problems() []error {
	var ret []error
	if match, _ := regexp.MatchString(`^https?$`, cc.Request.Protocol); !match {
		ret = append(ret, fmt.Errorf("invalid protocol: %s", cc.Request.Protocol))
	}
	if match, _ := regexp.MatchString(`\{\{\ *.Latitude *\}\}`, cc.Request.PathTemplate); !match {
		ret = append(ret, errors.New("missing {{.Latitude}} in path template"))
	}
	if match, _ := regexp.MatchString(`\{\{\ *.Longitude *\}\}`, cc.Request.PathTemplate); !match {
		ret = append(ret, errors.New("missing {{.Longitude}} in path template"))
	}
	if cc.Probe.MaxFailedLocations >= len(cc.Probe.Locations) {
		ret = append(ret, errors.New("max failures cannot be larger than length of check locations"))
	}
	return ret
}

// GetRequests returns a list of all possible permutations of server address
// and lat/lon.
func (cc *ProbeConfiguration) GetRequests() []NamedRequest {
	var ret []NamedRequest

	for _, server := range cc.Request.Servers {
		base := fmt.Sprintf("%s://%s%s", cc.Request.Protocol, server, cc.Request.PathTemplate)
		tmpl, err := template.New("CheckConfiguration").Parse(base)
		if err != nil {
			panic(err)
		}
		for _, loc := range cc.Probe.Locations {
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

type NamedRequest struct {
	Name      string
	URL       *url.URL
	Blueprint Blueprint
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
	MaxAge     Duration                `json:"max_age"`
	Timeseries Timeseries              `json:"timeseries"`
	Data       map[string]TimestepData `json:"data"`
}

type Timeseries struct {
	Timeresolution []Duration `json:"timeresolution"`
	MinDuration    Duration   `json:"minduration"`
}

type TimestepData struct {
	Summary map[string]CheckSpecification `json:"summary"`
	Details map[string]CheckSpecification `json:"details"`
}

type Parameter struct {
	Name string             `json:"name"`
	Spec CheckSpecification `json:"spec"`
}

type CheckSpecification struct {
	MinimumCount int     `json:"min_count,omitempty"`
	MinimumValue float32 `json:"min,omitempty"`
	MaximumValue float32 `json:"max,omitempty"`
	Optional     bool    `json:"optional,omitempty"`
}

// Duration wraps time.Duration so we can get the correct type directly from json unmarshalling
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
