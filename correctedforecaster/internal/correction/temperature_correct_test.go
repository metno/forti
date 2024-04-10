package correction

import (
	"math"
	"testing"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/parameters/weathersymbol"
)

func TestDewPointTemperature(t *testing.T) {
	result := dewPointTemperature(.60060547, 17.02835)
	expected := 9.250092
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("unexpected dew point temperature: %v", result)
	}
}

func TestWeatherSymbol6h(t *testing.T) {
	const timeSteps = 4

	var times []time.Time
	d := time.Date(2024, 4, 9, 0, 0, 0, 0, time.UTC)
	for h := time.Duration(0); h < timeSteps*time.Hour; h += time.Hour {
		times = append(times, d.Add(h))
	}

	forecast := map[string]internalprotocol.InterpretedData{
		"weather_symbol_6h": {
			Times:  times,
			Values: repeat(weathersymbol.Snow.ToValue(), timeSteps),
		},
		"air_temperature_2m_min6h": {
			Times:  times,
			Values: repeat(7, timeSteps),
		},
		"air_temperature_2m_max6h": {
			Times:  times,
			Values: repeat(9, timeSteps),
		},
	}

	UpdateSymbols6h(forecast)

	symbols := forecast["weather_symbol_6h"]
	for i := 0; i < timeSteps; i++ {
		symbol := weathersymbol.WeatherSymbol(symbols.Values[i])
		if symbol != weathersymbol.Rain {
			t.Errorf("expected %s for weather symbol at timestep %d, got %s", weathersymbol.Rain, i, symbol)
		}
	}
}

func repeat(value float32, times int) []float32 {
	ret := make([]float32, times)
	for i := 0; i < times; i++ {
		ret[i] = value
	}
	return ret
}
