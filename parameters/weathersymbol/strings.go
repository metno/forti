package weathersymbol

var pretty = map[WeatherSymbol]string{
	0:  "<none>",
	1:  "Clear sky",     // (d/n/m)
	2:  "Fair",          // (d/n/m)
	3:  "Partly cloudy", // (d/n/m)
	4:  "Cloudy",
	40: "Light rain showers",               // (d/n/m)
	5:  "Rain showers",                     // (d/n/m)
	41: "Heavy rain showers",               // (d/n/m)
	24: "Light rain showers and thunder",   // (d/n/m)
	6:  "Rain showers and thunder",         // (d/n/m)
	25: "Heavy rain showers and thunder",   // (d/n/m)
	42: "Light sleet showers",              // (d/n/m)
	7:  "Sleet showers",                    // (d/n/m)
	43: "Heavy sleet showers",              // (d/n/m)
	26: "Lights sleet showers and thunder", // (d/n/m)
	20: "Sleet showers and thunder",        // (d/n/m)
	27: "Heavy sleet showers and thunder",  // (d/n/m)
	44: "Light snow showers",               // (d/n/m)
	8:  "Snow showers",                     // (d/n/m)
	45: "Heavy snow showers",               // (d/n/m)
	28: "Lights snow showers and thunder",  // (d/n/m)
	21: "Snow showers and thunder",         // (d/n/m)
	29: "Heavy snow showers and thunder",   // (d/n/m)
	46: "Light rain",
	9:  "Rain",
	10: "Heavy rain",
	30: "Light rain and thunder",
	22: "Rain and thunder",
	11: "Heavy rain and thunder",
	47: "Light sleet",
	12: "Sleet",
	48: "Heavy sleet",
	31: "Light sleet and thunder",
	23: "Sleet and thunder",
	32: "Heavy sleet and thunder",
	49: "Light snow",
	13: "Snow",
	50: "Heavy snow",
	33: "Light snow and thunder",
	14: "Snow and thunder",
	34: "Heavy snow and thunder",
	15: "Fog",
}
