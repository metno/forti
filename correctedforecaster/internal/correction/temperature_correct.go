package correction

import (
	"math"
	"strings"

	"gitlab.met.no/forti/f2/simpleforecaster/pkg/forecaster"
	"gitlab.met.no/forti/f2/weathersymbol"
)

// UpdateTemperature performs corrections on air temperature variables, based
// on the given difference in altitude.
func UpdateTemperature(interpreted map[string]forecaster.InterpretedData, altitudeDiff float32) {
	correction := 0.006 * altitudeDiff

	for parameter, data := range interpreted {
		if isTemperature(parameter) {
			for i := range data.Values {
				data.Values[i] += correction
			}
		}
	}
}

// UpdateSymbols modifies a symbol to take into account any changes in
// temperature after the symbol was generated.
func UpdateSymbols(interpreted map[string]forecaster.InterpretedData) {
	symbols, ok := interpreted["weather_symbol"]
	if !ok {
		return
	}
	instant := "air_temperature_2m"

	sorted := forecaster.SortByTime(interpreted, instant)

	for i, t := range symbols.Times {
		values, ok := sorted[t]
		if !ok {
			// no temperature data related to symbol
			continue
		}
		t, ok := values[instant]
		if ok {
			symbol := weathersymbol.FromValue(symbols.Values[i])
			symbol = symbol.WithCorrectedTemperature(t)
			symbols.Values[i] = symbol.ToValue()
		}
	}
}

// UpdateSymbols6h modifies 6-hourly weather symbols, in the same way as
// UpdateSymbols does.
func UpdateSymbols6h(interpreted map[string]forecaster.InterpretedData) {
	symbols, ok := interpreted["weather_symbol_6h"]
	if !ok {
		return
	}
	min := "air_temperature_2m_min6h"
	max := "air_temperature_2m_max6h"
	instant := "air_temperature_2m"

	sorted := forecaster.SortByTime(interpreted, min, max, instant)

	for i, t := range symbols.Times {
		values, ok := sorted[t]
		if !ok {
			// no temperature data related to symbol
			continue
		}

		a, aok := values[min]
		b, bok := values[max]
		if aok && bok {
			symbol := weathersymbol.FromValue(symbols.Values[i])
			symbol = symbol.WithCorrectedTemperature((a + b) / 2)
			symbols.Values[i] = symbol.ToValue()
		} else {
			t, ok := values[instant]
			if ok {
				symbol := weathersymbol.FromValue(symbols.Values[i])
				symbol = symbol.WithCorrectedTemperature(t)
				symbols.Values[i] = symbol.ToValue()
			}
		}
	}
}

// UpdateDewpointTemperature recalculates dew point temperature, based on
// relative humidity and air temperature.
func UpdateDewpointTemperature(interpreted map[string]forecaster.InterpretedData) {
	dewPoint, ok := interpreted["dew_point_temperature_2m"]
	if !ok {
		return
	}

	temperature := "air_temperature_2m"
	humidity := "relative_humidity_2m"
	sorted := forecaster.SortByTime(interpreted, humidity, temperature)

	for i, t := range dewPoint.Times {
		timestep, ok := sorted[t]
		if !ok {
			continue
		}
		h, ok := timestep[humidity]
		if !ok {
			continue
		}
		t, ok := timestep[temperature]
		if !ok {
			continue
		}
		dewPoint.Values[i] = float32(dewPointTemperature(float64(h)/100, float64(t)))
	}
}

// dewPointTemperature calculates dew point temperature from the given
// relative humidity (range 0-1) and temperature (in celsius).
func dewPointTemperature(humidity, temperature float64) float64 {
	e := humidity * 0.611 * math.Exp((17.63*temperature)/(temperature+243.04))
	dewPoint := (116.9 + 243.04*math.Log(e)) / (16.78 - math.Log(e))

	if dewPoint > temperature {
		dewPoint = temperature
	}

	return dewPoint
}

// isTemperature determins if the given parameter name represents a temperature.
func isTemperature(param string) bool {
	return strings.HasPrefix(param, "air_temperature") && !strings.HasSuffix(param, "_code")
}
