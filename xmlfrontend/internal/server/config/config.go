package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Configuration *Elements

func Initialize(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var conf Elements
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return fmt.Errorf("unable to read json configuration: %w", err)
	}
	Configuration = &conf

	return nil
}

type Elements struct {
	CutForecast bool          `json:"cut_forecast,omitempty"`
	HTTPHeaders []HTTPHeader  `json:"http_headers"`
	Elements    []DataElement `json:"elements"`
}

type HTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DataElement struct {
	Offset     string      `json:"offset"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name          string          `json:"name"`
	NcName        string          `json:"ncName"`
	ValueName     string          `json:"valueName"`
	Attrs         []Attrs         `json:"attrs,omitempty"`
	ComputedAttrs []ComputedAttrs `json:"computedAttrs,omitempty"`
}

type Attrs struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ComputedAttrs struct {
	Name string `json:"name"`
	Func string `json:"func"`
}
