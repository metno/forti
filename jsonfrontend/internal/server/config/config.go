package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

var Configuration *Elements

// InitializeFromString populates global Configuration with values from string.
func InitializeFromString(config string) error {
	br := bytes.NewReader([]byte(config))

	var conf Elements
	if err := json.NewDecoder(br).Decode(&conf); err != nil {
		return fmt.Errorf("unable to read json configuration: %w", err)
	}
	Configuration = &conf

	return nil
}

// Initialize populates global Configuration with values from specified file path.
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
	CutForecast      bool                 `json:"cut_forecast,omitempty"`
	LocationFromGrid bool                 `json:"location_from_grid,omitempty"`
	HTTPHeaders      []HTTPHeader         `json:"http_headers"`
	OfferGzip        bool                 `json:"offer_gzip"`
	SkipAltitude     bool                 `json:"skip_altitude,omitempty"`
	DataExpiryOffset int                  `json:"data_expiry_offset"`
	Meta             Meta                 `json:"meta,omitempty"`
	Parameters       map[string]TimeGroup `json:"parameters"`
}

type HTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Meta struct {
	RadarCoverage string `json:"radar_coverage,omitempty"`
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
