package weathersymbol

const (
	None                        WeatherSymbol = 0
	ClearSky                    WeatherSymbol = 1
	Fair                        WeatherSymbol = 2
	PartlyCloudy                WeatherSymbol = 3
	Cloudy                      WeatherSymbol = 4
	LightRainShowers            WeatherSymbol = 40
	RainShowers                 WeatherSymbol = 5
	HeavyRainShowers            WeatherSymbol = 41
	LightRainShowersAndThunder  WeatherSymbol = 24
	RainShowersAndThunder       WeatherSymbol = 6
	HeavyRainShowersAndThunder  WeatherSymbol = 25
	LightSleetShowers           WeatherSymbol = 42
	SleetShowers                WeatherSymbol = 7
	HeavySleetShowers           WeatherSymbol = 43
	LightSleetShowersAndThunder WeatherSymbol = 26
	SleetShowersAndThunder      WeatherSymbol = 20
	HeavySleetShowersAndThunder WeatherSymbol = 27
	LightSnowShowers            WeatherSymbol = 44
	SnowShowers                 WeatherSymbol = 8
	HeavySnowShowers            WeatherSymbol = 45
	LightSnowShowersAndThunder  WeatherSymbol = 28
	SnowShowersAndThunder       WeatherSymbol = 21
	HeavySnowShowersAndThunder  WeatherSymbol = 29
	LightRain                   WeatherSymbol = 46
	Rain                        WeatherSymbol = 9
	HeavyRain                   WeatherSymbol = 10
	LightRainAndThunder         WeatherSymbol = 30
	RainAndThunder              WeatherSymbol = 22
	HeavyRainAndThunder         WeatherSymbol = 11
	LightSleet                  WeatherSymbol = 47
	Sleet                       WeatherSymbol = 12
	HeavySleet                  WeatherSymbol = 48
	LightSleetAndThunder        WeatherSymbol = 31
	SleetAndThunder             WeatherSymbol = 23
	HeavySleetAndThunder        WeatherSymbol = 32
	LightSnow                   WeatherSymbol = 49
	Snow                        WeatherSymbol = 13
	HeavySnow                   WeatherSymbol = 50
	LightSnowAndThunder         WeatherSymbol = 33
	SnowAndThunder              WeatherSymbol = 14
	HeavySnowAndThunder         WeatherSymbol = 34
	Fog                         WeatherSymbol = 15

	// SymbolMask Can Be Used To Remove Nighttime Information From A Weather Symbol

	// // Night Can Be Or'ed Into A Weather Symbol To Indicate Nighttime
	// Night = 1 << 7

	// // PolarTwighlight Can Be Or'ed Into A Weather Symbol To Indicate Polar Twighlight
	// PolarTwighlight = 1 << 8
)

// SunState specifies if the sun is up or down.
type SunState int

const (
	// Up means that the sun is above the horizon
	Up SunState = 0

	// Down means that the sun is below the horizon
	Down SunState = 1 << 7

	// PolarTwighlight is only used in arctic areas, during the time when the
	// sun is always below the horizon.
	// It is used when the sun is so close to the horizon that the twighlight
	// will be perceived as daylight.
	PolarTwighlight SunState = 1 << 8
)

// symbolMask Can Be Used To Remove Nighttime Information From A Weather Symbol
const sunStateMask = ^WeatherSymbol(Up | Down | PolarTwighlight)

func (s SunState) String() string {
	switch s {
	case Up:
		return "day"
	case Down:
		return "night"
	case PolarTwighlight:
		return "polartwilight"
	}
	return "<error>"
}
