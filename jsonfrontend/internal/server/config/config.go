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
	CutForecast bool                 `json:"cut_forecast,omitempty"`
	HTTPHeaders []HTTPHeader         `json:"http_headers"`
	Parameters  map[string]TimeGroup `json:"parameters"`
}

type HTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TimeGroup struct {
	Offset     int               `json:"offset"`
	Summary    *Summary          `json:"summary,omitempty"`
	Parameters map[string]string `json:"parameters"`
}

type Summary struct {
	SymbolCode       string `json:"symbol_code"`
	SymbolConfidence string `json:"symbol_confidence,omitmepty"`
}
