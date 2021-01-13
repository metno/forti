package weathersymbol

import "fmt"

// WeatherSymbol represents a summary of the weather at a particular time
type WeatherSymbol int

// FromValue creates a new weather symbol from the given float value. No sanity
// check is done on the returned symbol.
func FromValue(val float32) WeatherSymbol {
	symbol := WeatherSymbol(val)
	return symbol
}

// ToValue returns the integer value for this symbol
func (w WeatherSymbol) ToValue() float32 {
	return float32(w)
}

// IsValid tells if this weather symbol has a sane value.
func (w WeatherSymbol) IsValid() bool {
	_, ok := identifiers[w.WithoutSunState()]
	return ok
}

// String returns a human-readable text describing the weather symobl.
func (w WeatherSymbol) String() string {
	str, ok := pretty[w.WithoutSunState()]
	if !ok {
		return "<error>"
	}

	return str
}

// Identifier returns a unique identifier for the weather symbol, or the
// special string "<error>" if the symbol is invalid.
func (w WeatherSymbol) Identifier() string {

	str, ok := identifiers[w.WithoutSunState()]
	if !ok {
		return "<error>"
	}

	if withSun := identifiersWithSun[w.WithoutSunState()]; !withSun {
		return str
	}

	return fmt.Sprintf("%s_%s", str, w.SunState())
}

// WithSun returns a new symbol with the given sun symbol.
func (w WeatherSymbol) WithSun(p SunState) WeatherSymbol {
	w &= sunStateMask
	return w | WeatherSymbol(p)
}

// WithoutSunState removes sun state from the weather symbol
func (w WeatherSymbol) WithoutSunState() WeatherSymbol {
	return w & sunStateMask
}

// SunState extracts sun state from symbol
func (w WeatherSymbol) SunState() SunState {
	return SunState(w) & (Up | Down | PolarTwighlight)
}

// WithCorrectedTemperature changes the symbol precipitation phase (rain,
// sleet snow) into another one, according to the given temperature.
func (w WeatherSymbol) WithCorrectedTemperature(temperature float32) WeatherSymbol {
	phase := PhaseRain
	if temperature <= 0.5 {
		phase = PhaseSnow
	} else if temperature < 2 {
		phase = PhaseSleet
	}
	return w.WithPhase(phase)
}

// WithPhase modifies the eather symbol to use the given precipitation phase.
func (w WeatherSymbol) WithPhase(p PrecipitationPhase) WeatherSymbol {

	value := WeatherSymbol(w & sunStateMask)
	dayNight := SunState(w) & (Down | PolarTwighlight)

	transitionMap := phaseTransitions[p]
	s, ok := transitionMap[value]
	if !ok {
		return w
	}

	return s.WithSun(dayNight)
}
