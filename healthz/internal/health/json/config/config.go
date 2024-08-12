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
	setDefaultCheckWindow(&conf)

	return &conf, nil
}

// setDefaultWindow sets default check window if size is zero or negative.
// default window settings will accept 1 failure in 10 checks.
func setDefaultCheckWindow(conf *CheckConfiguration) {
	if conf.CheckWindow.Size < 1 {
		conf.CheckWindow.Size = 10
		conf.CheckWindow.FailThreshold = 1
	}
}

// CheckConfiguration contains a spec for how to execute sanity checks on
// various locationforecast servers- and locations.
type CheckConfiguration struct {
	Headers     map[string]string `json:"headers"`
	CheckWindow CheckWindow       `json:"check_window"`
	Request     Request           `json:"request"`
	Response    Response          `json:"response"`
}

type Request struct {
	Protocol     string   `json:"protocol"`
	Servers      []string `json:"servers"`
	PathTemplate string   `json:"path_template"`
}

// Size is the number of the most recent checks that are kept in memory.
// If more than FailThreshold of these checks have failed, the system is considered unhealthy.
type CheckWindow struct {
	Size          int `json:"size"`
	FailThreshold int `json:"fail_threshold"`
}

type Response struct {
	MaxFailures int        `json:"max_failures"`
	Locations   []Location `json:"locations"`
}

type NamedRequest struct {
	Name      string
	URL       *url.URL
	Blueprint Blueprint
}

// Problems returns a list of errors in the configuration, but only errors
// that are so severe that checks cannot be based on this config.
func (cc *CheckConfiguration) Problems() []error {
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
	if cc.Response.MaxFailures >= len(cc.Response.Locations) {
		ret = append(ret, errors.New("max failures cannot be larger than length of check locations"))
	}
	return ret
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
