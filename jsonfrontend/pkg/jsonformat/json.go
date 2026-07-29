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
	UpdatedAt     time.Time         `json:"updated_at"               jsonschema:"description=Time of the last forecast update"`
	Error         string            `json:"error,omitempty"          jsonschema:"description=Error message if the forecast is unavailable"`
	Units         map[string]string `json:"units"                    jsonschema:"description=Maps each forecast parameter name to its physical unit"`
	RadarCoverage string            `json:"radar_coverage,omitempty" jsonschema:"description=Custom metadata used only for timeseries forecasts from radar data. Radar coverage status for the location"`
}

type TimeStep struct {
	Time time.Time               `json:"time" jsonschema:"description=Start time of the forecast interval"`
	Data map[string]TimestepData `json:"data" jsonschema:"description=Forecast data keyed by interval length e.g. instance, next_1_hours"`
}

type TimestepData struct {
	Summary *Summary        `json:"summary,omitempty" jsonschema:"description=Summary of the forecast parameters"`
	Details ForecastDetails `json:"details" jsonschema:"description=List of forecast parameters and their numerical values"`
}

type Summary struct {
	SymbolCode       string `json:"symbol_code"                jsonschema:"description=Weather symbol identifier"`
	SymbolConfidence string `json:"symbol_confidence,omitempty" jsonschema:"description=Confidence level of the weather symbol"`
}

type ForecastDetails map[string]SingleDigitFloat

type SingleDigitFloat float32

func (f SingleDigitFloat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', 1, 32)), nil
}
