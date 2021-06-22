package jsonformat

import (
	"strconv"
	"time"
)

type Forecast struct {
	Meta       Metadata   `json:"meta"`
	Timeseries []TimeStep `json:"timeseries"`
}

type Metadata struct {
	UpdatedAt     time.Time         `json:"updated_at"`
	Error         string            `json:"error,omitempty"`
	Units         map[string]string `json:"units"`
	RadarCoverage string            `json:"radar_coverage,omitempty"`
}

type TimeStep struct {
	Time time.Time               `json:"time"`
	Data map[string]TimestepData `json:"data"`
}

type TimestepData struct {
	Summary *Summary        `json:"summary,omitempty"`
	Details ForecastDetails `json:"details,omitempty"`
}

type Summary struct {
	SymbolCode       string `json:"symbol_code"`
	SymbolConfidence string `json:"symbol_confidence,omitempty"`
}

type ForecastDetails map[string]SingleDigitFloat

type SingleDigitFloat float32

func (f SingleDigitFloat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', 1, 32)), nil
}
