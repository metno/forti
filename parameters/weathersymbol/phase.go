package weathersymbol

// PrecipitationPhase specifies what form the water in a rainfall takes.
type PrecipitationPhase int

const (
	// PhaseRain is precipitation as liquid water
	PhaseRain PrecipitationPhase = iota

	// PhaseSleet is precipitation as a mix between water and snow
	PhaseSleet

	// PhaseSnow is precipitation as snow
	PhaseSnow
)

func (p PrecipitationPhase) String() string {
	switch p {
	case PhaseRain:
		return "rain"
	case PhaseSleet:
		return "sleet"
	case PhaseSnow:
		return "snow"
	default:
		return "???"
	}
}

var phaseTransitions = map[PrecipitationPhase]map[WeatherSymbol]WeatherSymbol{
	PhaseRain: {
		// Sleet
		LightSleet:                  LightRain,
		Sleet:                       Rain,
		HeavySleet:                  HeavyRain,
		LightSleetShowers:           LightRainShowers,
		SleetShowers:                RainShowers,
		HeavySleetShowers:           HeavyRainShowers,
		LightSleetAndThunder:        LightRainAndThunder,
		SleetAndThunder:             RainShowersAndThunder,
		HeavySleetAndThunder:        HeavyRainAndThunder,
		LightSleetShowersAndThunder: LightRainShowersAndThunder,
		SleetShowersAndThunder:      RainShowersAndThunder,
		HeavySleetShowersAndThunder: HeavyRainShowersAndThunder,

		// Snow
		LightSnow:                  LightRain,
		Snow:                       Rain,
		HeavySnow:                  HeavyRain,
		LightSnowShowers:           LightRainShowers,
		SnowShowers:                RainShowers,
		HeavySnowShowers:           HeavyRainShowers,
		LightSnowAndThunder:        LightRainAndThunder,
		SnowAndThunder:             RainShowersAndThunder,
		HeavySnowAndThunder:        HeavyRainAndThunder,
		LightSnowShowersAndThunder: LightRainShowersAndThunder,
		SnowShowersAndThunder:      RainShowersAndThunder,
		HeavySnowShowersAndThunder: HeavyRainShowersAndThunder,
	},
	PhaseSleet: {

		// Rain
		LightRain:                  LightSleet,
		Rain:                       Sleet,
		HeavyRain:                  HeavySleet,
		LightRainShowers:           LightSleetShowers,
		RainShowers:                SleetShowers,
		HeavyRainShowers:           HeavySleetShowers,
		LightRainAndThunder:        LightSleetAndThunder,
		RainAndThunder:             SleetAndThunder,
		HeavyRainAndThunder:        HeavySleetAndThunder,
		LightRainShowersAndThunder: LightSleetShowersAndThunder,
		RainShowersAndThunder:      SleetShowersAndThunder,
		HeavyRainShowersAndThunder: HeavySleetShowersAndThunder,

		// Snow
		LightSnow:                  LightSleet,
		Snow:                       Sleet,
		HeavySnow:                  HeavySleet,
		LightSnowShowers:           LightSleetShowers,
		SnowShowers:                SleetShowers,
		HeavySnowShowers:           HeavySleetShowers,
		LightSnowAndThunder:        LightSleetAndThunder,
		SnowAndThunder:             SleetAndThunder,
		HeavySnowAndThunder:        HeavySleetAndThunder,
		LightSnowShowersAndThunder: LightSleetShowersAndThunder,
		SnowShowersAndThunder:      SleetShowersAndThunder,
		HeavySnowShowersAndThunder: HeavySleetShowersAndThunder,
	},
	PhaseSnow: {

		// Rain
		LightRain:                  LightSnow,
		Rain:                       Snow,
		HeavyRain:                  HeavySnow,
		LightRainShowers:           LightSnowShowers,
		RainShowers:                SnowShowers,
		HeavyRainShowers:           HeavySnowShowers,
		LightRainAndThunder:        LightSnowAndThunder,
		RainAndThunder:             SnowAndThunder,
		HeavyRainAndThunder:        HeavySnowAndThunder,
		LightRainShowersAndThunder: LightSnowShowersAndThunder,
		RainShowersAndThunder:      SnowShowersAndThunder,
		HeavyRainShowersAndThunder: HeavySnowShowersAndThunder,

		// Sleet
		LightSleet:                  LightSnow,
		Sleet:                       Snow,
		HeavySleet:                  HeavySnow,
		LightSleetShowers:           LightSnowShowers,
		SleetShowers:                SnowShowers,
		HeavySleetShowers:           HeavySnowShowers,
		LightSleetAndThunder:        LightSnowAndThunder,
		SleetAndThunder:             SnowAndThunder,
		HeavySleetAndThunder:        HeavySnowAndThunder,
		LightSleetShowersAndThunder: LightSnowShowersAndThunder,
		SleetShowersAndThunder:      SnowShowersAndThunder,
		HeavySleetShowersAndThunder: HeavySnowShowersAndThunder,
	},
}
