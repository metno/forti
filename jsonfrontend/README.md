# jsonfrontend

REST API frontend that serves weather forecast data in JSON format (GeoJSON). This is the recommended frontend for all new integrations.

See the [API Reference](../docs/api-reference.md) for endpoint documentation.

## Usage

See `jsonfrontend --help` for command-line options.

The service listens on `:8080` for forecast requests.

## Configuration

The configuration file (`jsonformat.json`) controls how internal forecast parameters are mapped to external API responses, which time periods are exposed, and HTTP response behavior.

### Minimal Example

```json
{
  "parameters": {
    "instant": {
      "offset": 0,
      "parameters": {
        "air_temperature_2m": "air_temperature",
        "wind_speed_10m": "wind_speed"
      }
    },
    "next_6_hours": {
      "offset": 0,
      "summary": {
        "symbol_code": "weather_symbol_6h"
      },
      "parameters": {
        "precipitation_amount_6h": "precipitation_amount"
      }
    }
  },
  "units": {},
  "http_headers": [],
  "offer_gzip": true,
  "data_expiry_offset": 180
}
```

### Configuration Options

#### `parameters`

Maps internal forecast parameters to external API names and organizes them by time period.

- **Key**: Time period identifier (e.g., `"instant"`, `"next_1_hours"`, `"next_6_hours"`, `"next_12_hours"`)
- **Value**: Time period configuration object (see below)

Each time period appears as a separate section in the JSON response under `timeseries[].data.<period>`.

**Time Period Object:**

- **`offset`** (integer, required): Time offset in hours. Typically `0` for instant parameters, or any other value for "next_X_hours" parameters that describe the period starting at the timestep.
- **`parameters`** (object, required): Maps internal parameter names (keys) to external API names (values).
  - Example: `{"air_temperature_2m": "air_temperature"}` means the internal parameter `air_temperature_2m` is exposed as `air_temperature` in the API.
- **`summary`** (object, optional): Defines weather summary symbols for this period (only for period parameters like `next_1_hours`).
  - **`symbol_code`** (string): Internal parameter name containing the weather symbol code
  - **`symbol_confidence`** (string, optional): Internal parameter name for symbol confidence

**Example with multiple periods:**

```json
{
  "parameters": {
    "instant": {
      "offset": 0,
      "parameters": {
        "air_temperature_2m": "air_temperature",
        "relative_humidity_2m": "relative_humidity",
        "wind_speed_10m": "wind_speed",
        "wind_direction_10m": "wind_from_direction",
        "air_pressure_at_sea_level": "air_pressure_at_sea_level",
        "cloud_area_fraction": "cloud_area_fraction"
      }
    },
    "next_1_hours": {
      "offset": 0,
      "summary": {
        "symbol_code": "weather_symbol_1h",
        "symbol_confidence": "symbol_confidence"
      },
      "parameters": {
        "precipitation_amount_1h": "precipitation_amount"
      }
    },
    "next_6_hours": {
      "offset": 0,
      "summary": {
        "symbol_code": "weather_symbol_6h"
      },
      "parameters": {
        "precipitation_amount_6h": "precipitation_amount",
        "air_temperature_max": "air_temperature_max",
        "air_temperature_min": "air_temperature_min"
      }
    }
  }
}
```

#### `http_headers`

Custom HTTP response headers to include in every response.

```json
{
  "http_headers": [
    {
      "key": "Access-Control-Allow-Origin",
      "value": "*"
    }
  ]
}
```

- **`key`** (string): HTTP header name
- **`value`** (string): HTTP header value

#### `offer_gzip`

- **`offer_gzip`** (boolean, optional): If `true`, enables gzip compression for responses when the client sends `Accept-Encoding: gzip`. Default: `false`.

#### `data_expiry_offset`

- **`data_expiry_offset`** (integer, optional): Maximum cache lifetime in seconds for the `Expires` HTTP header. The actual expiry time is randomized between 80% and 100% of this value to prevent cache stampedes. If `<= 0`, uses 60 seconds. Default: `0`.

#### `cut_forecast`

- **`cut_forecast`** (boolean, optional): If `true`, removes timesteps in the past (before the current hour). Useful for reducing response size when historical data isn't needed. Default: `false`.

#### `skip_altitude`

- **`skip_altitude`** (boolean, optional): If `true`, omits altitude from the geometry coordinates in the response. Default: `false`.

#### `location_from_grid`

- **`location_from_grid`** (boolean, optional): If `true`, uses the actual grid cell location instead of the requested lat/lon in the response geometry. Default: `false`.

#### `meta`

Metadata configuration for special parameters.

```json
{
  "meta": {
    "radar_coverage": "radar_coverage_internal_param"
  }
}
```

- **`radar_coverage`** (string, optional): Internal parameter name for radar coverage status. If present, this value is included in the response metadata as `radar_coverage`.

## Parameter Naming

Internal parameter names depend on your data source and are defined by the forecast datasets. The configuration file maps these internal names to the external API names that clients see.

For consistency with MET Norway's locationforecast API, use the external parameter names documented in the [MET Norway data model](https://api.met.no/doc/locationforecast/datamodel).

## Units

Units for each parameter are extracted automatically from the forecast data metadata. The configuration file does not need to specify units unless you want to override them.
