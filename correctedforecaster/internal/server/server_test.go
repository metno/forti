package server

import (
	"testing"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/parameters/weathersymbol"
)

func TestCorrectWithBetterTopography(t *testing.T) {
	var s Server
	request := internalprotocol.GetForecastRequest{
		Altitude: &internalprotocol.Altitude{
			Value:    1000,
			Override: true,
		},
	}

	const timeSteps = 25

	var times []time.Time
	d := time.Date(2024, 4, 9, 0, 0, 0, 0, time.UTC)
	for h := time.Hour; h < timeSteps*time.Hour; h += time.Hour {
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
			Times:  times,
			Values: repeat(float32(weathersymbol.Rain), timeSteps),
		},
		"weather_symbol_12h": {
			Times:  times,
			Values: repeat(float32(weathersymbol.Rain), timeSteps),
		},
		"air_temperature_2m": {
			Times:  times,
			Values: repeat(2, timeSteps),
		},
		"air_temperature_2m_min6h": {
			Times:  times,
			Values: repeat(2, timeSteps),
		},
		"air_temperature_2m_max6h": {
			Times:  times,
			Values: repeat(2, timeSteps),
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

	for i := 0; i < timeSteps-1; i++ {
		symbol := weathersymbol.WeatherSymbol(interpreted["weather_symbol"].Values[i])
		if symbol != weathersymbol.Snow {
			t.Errorf("expected %s for 1h weather symbol at timestep %d, got %s", weathersymbol.Snow, i, symbol)
		}
	}
	for i := 0; i < timeSteps-1; i++ {
		symbol := weathersymbol.WeatherSymbol(interpreted["weather_symbol_6h"].Values[i])
		if symbol != weathersymbol.Snow {
			t.Errorf("expected %s for 1h weather symbol at timestep %d, got %s", weathersymbol.Snow, i, symbol)
		}
	}
	for i := 0; i < timeSteps; i++ {
		symbol := weathersymbol.WeatherSymbol(interpreted["weather_symbol_12h"].Values[i])
		if symbol != weathersymbol.Snow {
			t.Errorf("expected %s for 1h weather symbol at timestep %d, got %s", weathersymbol.Snow, i, symbol)
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
