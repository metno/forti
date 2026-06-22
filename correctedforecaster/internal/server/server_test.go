package server

import (
	"testing"
	"time"

	"github.com/metno/forti/internalprotocol"
	"github.com/metno/go-weathersymbol"
)

func TestCorrectWithBetterTopography(t *testing.T) {
	var s Server
	request := internalprotocol.GetForecastRequest{
		Altitude: &internalprotocol.Altitude{
			Value:    1000,
			Override: true,
		},
	}

	const timeSteps = 24

	var times []time.Time
	d := time.Date(2024, 4, 9, 0, 0, 0, 0, time.UTC)
	for h := time.Duration(0); h < timeSteps*time.Hour; h += time.Hour {
		times = append(times, d.Add(h))
	}

	interpreted := map[string]internalprotocol.InterpretedData{
		"altitude": {
			Times:  make([]time.Time, 1),
			Values: []float32{0},
		},
		"weather_symbol": {
			Times:  times,
			Values: repeat(float32(weathersymbol.Rain), timeSteps),
		},
		"weather_symbol_6h": {
			Times:  times[:timeSteps-6],
			Values: repeat(float32(weathersymbol.Rain), timeSteps-6),
		},
		"weather_symbol_12h": {
			Times:  times[:timeSteps-12],
			Values: repeat(float32(weathersymbol.Rain), timeSteps-12),
		},
		"air_temperature_2m": {
			Times:  times,
			Values: repeat(2, timeSteps),
		},
		"air_temperature_2m_min6h": {
			Times:  times[:timeSteps-6],
			Values: repeat(2, len(times)-6),
		},
		"air_temperature_2m_max6h": {
			Times:  times[:timeSteps-6],
			Values: repeat(2, len(times)-6),
		},
	}

	err := s.correctWithBetterTopography(&request, interpreted)
	if err != nil {
		t.Errorf("error when calling function correctWithBetterTopography: %v", err)
	}

	for i := 0; i < timeSteps; i++ {
		ta := interpreted["air_temperature_2m"].Values[i]
		if ta != -4 {
			t.Errorf("expected temperature %v at timestep %d, got %v", -4, i, ta)
		}
	}

	for i := 0; i < timeSteps; i++ {
		symbol := weathersymbol.WeatherSymbol(interpreted["weather_symbol"].Values[i])
		if symbol != weathersymbol.Snow {
			t.Errorf("expected %s for 1h weather symbol at timestep %d, got %s", weathersymbol.Snow, i, symbol)
		}
	}
	for i := 0; i < timeSteps-6; i++ {
		symbol := weathersymbol.WeatherSymbol(interpreted["weather_symbol_6h"].Values[i])
		if symbol != weathersymbol.Snow {
			t.Errorf("expected %s for 6h weather symbol at timestep %d, got %s", weathersymbol.Snow, i, symbol)
		}
	}
	for i := 0; i < timeSteps-12; i++ {
		symbol := weathersymbol.WeatherSymbol(interpreted["weather_symbol_12h"].Values[i])
		if symbol != weathersymbol.Snow {
			t.Errorf("expected %s for 12h weather symbol at timestep %d, got %s", weathersymbol.Snow, i, symbol)
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
