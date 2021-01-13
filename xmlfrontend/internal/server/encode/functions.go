package encode

import (
	"errors"
	"fmt"
	"strconv"

	"gitlab.met.no/forti/f2/parameters/weathersymbol"
)

var functions = map[string]func(value float32, allValues *map[string]float32) (string, error){
	"beaufort":              beaufort,
	"force_name":            forceName,
	"precipitation_high_1h": alternativeLookup("precipitation_amount_max1h"),
	"precipitation_high_6h": alternativeLookup("precipitation_amount_max6h"),
	"precipitation_low_1h":  alternativeLookup("precipitation_amount_min1h"),
	"precipitation_low_6h":  alternativeLookup("precipitation_amount_min6h"),
	"symbol_name":           symbolName,
	"symbol_value":          symbolValue,
	"symbol_string":         symbolString,
	"to_int":                toInt,
	"windDirection":         windDirection,
}

func beaufort(value float32, allValues *map[string]float32) (string, error) {
	if value < 0 {
		return "", fmt.Errorf("Invalid wind force: %fm/s", value)
	} else if value < 0.25 {
		return "0", nil
	} else if value < 1.55 {
		return "1", nil
	} else if value < 3.35 {
		return "2", nil
	} else if value < 5.45 {
		return "3", nil
	} else if value < 7.95 {
		return "4", nil
	} else if value < 10.75 {
		return "5", nil
	} else if value < 13.85 {
		return "6", nil
	} else if value < 17.15 {
		return "7", nil
	} else if value < 20.75 {
		return "8", nil
	} else if value < 24.45 {
		return "9", nil
	} else if value < 28.45 {
		return "10", nil
	} else if value < 32.65 {
		return "11", nil
	} else {
		return "12", nil
	}
}

func forceName(value float32, allValues *map[string]float32) (string, error) {
	b, err := beaufort(value, allValues)
	if err != nil {
		return "", err
	}
	switch b {
	case "0":
		return "Stille", nil
	case "1":
		return "Flau vind", nil
	case "2":
		return "Svak vind", nil
	case "3":
		return "Lett bris", nil
	case "4":
		return "Laber bris", nil
	case "5":
		return "Frisk bris", nil
	case "6":
		return "Liten kuling", nil
	case "7":
		return "Stiv kuling", nil
	case "8":
		return "Sterk kuling", nil
	case "9":
		return "Liten storm", nil
	case "10":
		return "Full storm", nil
	case "11":
		return "Sterk storm", nil
	case "12":
		return "Orkan", nil
	default:
		return "", fmt.Errorf("Unexpected value for beaufort: %s", b)
	}
}

func windDirection(value float32, allValues *map[string]float32) (string, error) {
	if value < 0 || value >= 360 {
		return "", fmt.Errorf("windDirection value of %f", value)
	} else if value < 22.5 {
		return "N", nil
	} else if value < 67.5 {
		return "NE", nil
	} else if value < 112.5 {
		return "E", nil
	} else if value < 157.5 {
		return "SE", nil
	} else if value < 202.5 {
		return "S", nil
	} else if value < 247.5 {
		return "SW", nil
	} else if value < 292.5 {
		return "W", nil
	} else if value < 337.5 {
		return "NW", nil
	} else {
		return "N", nil
	}
}

func symbolValue(value float32, allValues *map[string]float32) (string, error) {
	return fmt.Sprintf("%d", int(weathersymbol.FromValue(value).WithoutSunState())), nil
}

func symbolString(value float32, allValues *map[string]float32) (string, error) {
	symbol := weathersymbol.FromValue(value)
	return symbol.Identifier(), nil
}

func symbolName(value float32, allValues *map[string]float32) (string, error) {
	name, ok := symbols[int(weathersymbol.FromValue(value).WithoutSunState())]
	if !ok {
		return "Unknown", fmt.Errorf("unrecognized symbol %d", int(value))
	}
	return name, nil
}

func toInt(value float32, allValues *map[string]float32) (string, error) {
	return fmt.Sprintf("%d", int(value)), nil
}

func alternativeLookup(parameter string) func(value float32, allValues *map[string]float32) (string, error) {
	return func(value float32, allValues *map[string]float32) (string, error) {
		val, ok := (*allValues)[parameter]
		if !ok {
			return "", errors.New("did not find parameter " + parameter)
		}
		return strconv.FormatFloat(float64(val), 'f', 1, 32), nil
	}
}

func todo(value float32, allValues *map[string]float32) (string, error) {
	return "TODO", nil
}

var symbols map[int]string

func init() {
	symbols = map[int]string{
		1: "Sun", 101: "Dark_Sun",
		2: "LightCloud", 102: "Dark_LightCloud",
		3: "PartlyCloud", 103: "Dark_PartlyCloud",
		4: "Cloud",
		5: "LightRainSun", 105: "Dark_LightRainSun",
		6: "LightRainThunderSun", 106: "Dark_LightRainThunderSun",
		7: "SleetSun", 107: "Dark_SleetSun",
		8: "SnowSun", 108: "Dark_SnowSun",
		9:  "LightRain",
		10: "Rain",
		11: "RainThunder",
		12: "Sleet",
		13: "Snow",
		14: "SnowThunder",
		15: "Fog",
		20: "SleetSunThunder", 120: "Dark_SleetSunThunder",
		21: "SnowSunThunder", 121: "Dark_SnowSunThunder",
		22: "LightRainThunder",
		23: "SleetThunder",
		24: "DrizzleThunderSun", 124: "Dark_DrizzleThunderSun",
		25: "RainThunderSun", 125: "Dark_RainThunderSun",
		26: "LightSleetThunderSun", 126: "Dark_LightSleetThunderSun",
		27: "HeavySleetThunderSun", 127: "Dark_HeavySleetThunderSun",
		28: "LightSnowThunderSun", 128: "Dark_LightSnowThunderSun",
		29: "HeavySnowThunderSun", 129: "Dark_HeavySnowThunderSun",
		30: "DrizzleThunder",
		31: "LightSleetThunder",
		32: "HeavySleetThunder",
		33: "LightSnowThunder",
		34: "HeavySnowThunder",
		40: "DrizzleSun", 140: "Dark_DrizzleSun",
		41: "RainSun", 141: "Dark_RainSun",
		42: "LightSleetSun", 142: "Dark_LightSleetSun",
		43: "HeavySleetSun", 143: "Dark_HeavySleetSun",
		44: "LightSnowSun", 144: "Dark_LightSnowSun",
		45: "HeavySnowSun", 145: "Dark_HeavySnowSun",
		46: "Drizzle",
		47: "LightSleet",
		48: "HeavySleet",
		49: "LightSnow",
		50: "HeavySnow",
	}
}
