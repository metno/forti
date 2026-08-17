package correction

import (
	"math"
	"testing"
	"time"

	"github.com/metno/forti/internal/internalprotocol"
)

func TestDewPointTemperature(t *testing.T) {
	result := dewPointTemperature(.60060547, 17.02835)
	expected := 9.250092
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("unexpected dew point temperature: %v", result)
	}
}

func TestAirTemperatureCorrect(t *testing.T) {
	parametersToCorrect := []string{apparentAirTemperature, airTemperature2m, airTemperature2mMax6h}
	parametersToKeep := []string{"air_temperature_code", dewPointTemperature2m, "precipitation_amount"}

	times := []time.Time{time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC)}

	testingData := make(map[string]internalprotocol.InterpretedData)
	for _, param := range append(parametersToCorrect, parametersToKeep...) {
		testingData[param] = internalprotocol.InterpretedData{
			Times:  times,
			Values: []float32{20},
		}
	}

	UpdateTemperature(testingData, -1000)
	for _, p := range parametersToCorrect {
		const expected = 14.0
		corrected := testingData[p].Values[0]
		if corrected != expected {
			t.Errorf("%s: expected value %f, got %f", p, expected, corrected)
		}
	}
	for _, p := range parametersToKeep {
		const expected = 20.0
		corrected := testingData[p].Values[0]
		if corrected != expected {
			t.Errorf("%s: expected value %f, got %f", p, expected, corrected)
		}
	}
}
