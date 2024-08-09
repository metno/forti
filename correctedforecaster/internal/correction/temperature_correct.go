package correction

import (
	"math"
	"strings"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/parameters/weathersymbol"
)

// UpdateTemperature performs corrections on air temperature variables, based
// on the given difference in altitude.
func UpdateTemperature(interpreted map[string]internalprotocol.InterpretedData, altitudeDiff float32) {
	correction := 0.006 * altitudeDiff

	for parameter, data := range interpreted {
		if isTemperature(parameter) {
			for i := range data.Values {
				data.Values[i] += correction
			}
		}
	}
}

// UpdateSymbols changes all weather symbols in the given interpreted data to have the correct precipitation phase (snow, sleet, rain).
// It does this based on the temperatures found in the data.
func UpdateSymbols(interpreted map[string]internalprotocol.InterpretedData) {
	timesteps := internalprotocol.ReorganizeByTime(interpreted,
		weatherSymbol1h, weatherSymbol6h, weatherSymbol12h,
		airTemperature2m, airTemperature2mMin6h, airTemperature2mMax6h,
	)

	// Update 1h symbols
	if symbols, ok := interpreted[weatherSymbol1h]; ok {
		updateSymbols(symbols, timesteps, temperature1h)
	}

	// Update 6h symbols
	if symbols, ok := interpreted[weatherSymbol6h]; ok {
		updateSymbols(symbols, timesteps, avgTemperature6h)
	}

	// Update 12h symbols
	if symbols, ok := interpreted[weatherSymbol12h]; ok {
		updateSymbols(symbols, timesteps, avgTemperature12h)
	}
}

func updateSymbols(symbols internalprotocol.InterpretedData, timesteps map[time.Time]map[string]float32, getTemperature func(atTime time.Time, timesteps map[time.Time]map[string]float32) (value float32, ok bool)) {
	for i, t := range symbols.Times {
		temperature, ok := getTemperature(t, timesteps)
		if !ok {
			// Unable to determine a temperature for the symbol - we leave it unchanged.
			continue
		}
		symbol := weathersymbol.FromValue(symbols.Values[i])
		symbol = symbol.WithCorrectedTemperature(temperature)
		symbols.Values[i] = symbol.ToValue()
	}
}

func avgTemperature12h(atTime time.Time, timesteps map[time.Time]map[string]float32) (value float32, ok bool) {
	return avg(2, 6*time.Hour, atTime, timesteps, avgTemperature6h)
}

func avgTemperature6h(atTime time.Time, timesteps map[time.Time]map[string]float32) (value float32, ok bool) {
	timestep, ok := timesteps[atTime]
	if !ok {
		return float32(math.NaN()), false
	}

	minTemperature, ok := timestep[airTemperature2mMin6h]
	if !ok {
		return avg(6, time.Hour, atTime, timesteps, temperature1h)
	}
	maxTemperature, ok := timestep[airTemperature2mMax6h]
	if !ok {
		return avg(6, time.Hour, atTime, timesteps, temperature1h)
	}
	return (minTemperature + maxTemperature) / 2, true
}

func temperature1h(atTime time.Time, timesteps map[time.Time]map[string]float32) (value float32, ok bool) {
	timestep, ok := timesteps[atTime]
	if !ok {
		return float32(math.NaN()), false
	}
	value, ok = timestep[airTemperature2m]
	if !ok {
		return float32(math.NaN()), false
	}
	return
}

// avg reads values from <f>, by calling it <count> times.
// The time parameter to <f> at starts at <startTime>, and increases by <offset> each time <f> is called.
// The return value is the average of the values read.
func avg(
	count int,
	offset time.Duration,
	startTime time.Time,
	timesteps map[time.Time]map[string]float32,
	f func(atTime time.Time, timesteps map[time.Time]map[string]float32) (value float32, ok bool),
) (value float32, ok bool) {
	for i := 0; i < count; i++ {
		next, ok := f(startTime, timesteps)
		if !ok {
			return float32(math.NaN()), false
		}
		value += next
		startTime = startTime.Add(offset)
	}
	return value / float32(count), true
}

// UpdateDewpointTemperature recalculates dew point temperature, based on
// relative humidity and air temperature.
func UpdateDewpointTemperature(interpreted map[string]internalprotocol.InterpretedData) {
	dewPoint, ok := interpreted[dewPointTemperature2m]
	if !ok {
		return
	}

	temperature := airTemperature2m
	humidity := relativeHumidity2m
	sorted := internalprotocol.ReorganizeByTime(interpreted, humidity, temperature)

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
