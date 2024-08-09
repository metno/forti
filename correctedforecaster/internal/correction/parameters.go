package correction

// These are the expected parameter names from rawdataforecaster's grpc message

const (
	weatherSymbol1h       = "weather_symbol"
	weatherSymbol6h       = "weather_symbol_6h"
	weatherSymbol12h      = "weather_symbol_12h"
	airTemperature2m      = "air_temperature_2m"
	airTemperature2mMin6h = "air_temperature_2m_min6h"
	airTemperature2mMax6h = "air_temperature_2m_max6h"
	dewPointTemperature2m = "dew_point_temperature_2m"
	relativeHumidity2m    = "relative_humidity_2m"
)
