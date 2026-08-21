# API Reference

Forti provides a JSON REST API for retrieving weather and ocean forecast timeseries data for any geographical location. The API follows the [MET Norway locationforecast 2.0](https://api.met.no/weatherapi/locationforecast/2.0/documentation) format and structure.

## Endpoints

### GET /

Returns a point forecast for the specified location.

**Query Parameters:**

| Parameter | Type | Required | Range | Description |
|-----------|------|----------|-------|-------------|
| `lat` | float | Yes | -90 to 90 | Latitude in decimal degrees |
| `lon` | float | Yes | -180 to 180 | Longitude in decimal degrees |
| `altitude` | float | No | -500 to 9000 | Ground surface height above sea level in meters. When omitted, the internal topography model is used for temperature correction. |

**Example Request:**

```bash
curl 'http://localhost:8080/?lat=59.93&lon=10.72'
```

**Example with altitude:**

```bash
curl 'http://localhost:8080/?lat=59.93&lon=10.72&altitude=90'
```

## Response Format

The API returns forecast data in GeoJSON format with a structure compatible with MET Norway's [General Forecast JSON Format](https://api.met.no/doc/ForecastJSON).

### Success Response (200 OK)

```json
{
  "type": "Feature",
  "geometry": {
    "type": "Point",
    "coordinates": [10.72, 59.93, 90]
  },
  "properties": {
    "meta": {
      "updated_at": "2024-06-10T13:04:26Z",
      "units": {
        "air_temperature": "celsius",
        "wind_speed": "m/s",
        "precipitation_amount": "mm",
        ...
      }
    },
    "timeseries": [
      {
        "time": "2024-06-10T13:00:00Z",
        "data": {
          "instant": {
            "details": {
              "air_temperature": 20.7,
              "wind_speed": 2.5,
              "relative_humidity": 48.6,
              ...
            }
          },
          "next_1_hours": {
            "summary": {
              "symbol_code": "partlycloudy_day"
            },
            "details": {
              "precipitation_amount": 0.0
            }
          },
          "next_6_hours": {
            "summary": {
              "symbol_code": "partlycloudy_day"
            },
            "details": {
              "precipitation_amount": 0.2,
              "air_temperature_max": 22.1,
              "air_temperature_min": 18.4
            }
          }
        }
      },
      ...
    ]
  }
}
```

### Response Structure

**Geometry**
- `type`: Always "Feature" (GeoJSON)
- `geometry.type`: Always "Point"
- `geometry.coordinates`: Array of `[longitude, latitude, altitude]` (EPSG:4326)

**Properties**

The `properties` object contains two main sections:

#### Meta

| Field | Type | Description |
|-------|------|-------------|
| `updated_at` | ISO 8601 timestamp | When the forecast data was last updated |
| `units` | object | Units for all parameters in the forecast (e.g., `{"air_temperature": "celsius"}`) |
| `radar_coverage` | string | (Optional) Radar coverage status |

#### Timeseries

An array of forecast objects, each containing:

| Field | Type | Description |
|-------|------|-------------|
| `time` | ISO 8601 timestamp | Valid time for this forecast step (UTC) |
| `data` | object | Forecast data organized by time periods |

**Data Object Structure**

Each timeseries entry contains forecast data organized by time period:

- `instant`: Parameters valid at the specific timestamp
  - `details`: Key-value pairs of meteorological parameters (e.g., `air_temperature`, `wind_speed`)

- `next_1_hours`: Aggregated/summary data for the next 1 hour (short-range only)
  - `summary`: Weather summary (e.g., `symbol_code`)
  - `details`: Aggregated parameters (e.g., `precipitation_amount`)

- `next_6_hours`: Aggregated/summary data for the next 6 hours
  - `summary`: Weather summary
  - `details`: Aggregated parameters including min/max values

- `next_12_hours`: (May be present) Aggregated data for the next 12 hours

## Available Parameters

The specific parameters available depend on your Forti configuration and data sources. Common parameters include:

### Instant Parameters

These describe conditions at a specific point in time:

| Parameter | Unit | Description |
|-----------|------|-------------|
| `air_temperature` | celsius | Air temperature at 2m above ground |
| `air_pressure_at_sea_level` | hPa | Air pressure at sea level |
| `relative_humidity` | % | Relative humidity at 2m above ground |
| `wind_speed` | m/s | Wind speed at 10m above ground (10 min average) |
| `wind_from_direction` | degrees | Direction wind is coming from (0° = north, 90° = east) |
| `cloud_area_fraction` | % | Total cloud cover |
| `dew_point_temperature` | celsius | Dew point temperature at 2m above ground |
| `fog_area_fraction` | % | Fog coverage |
| `wind_speed_of_gust` | m/s | Maximum wind gust |

### Period Parameters

These describe aggregated values over a time period:

| Parameter | Unit | Description |
|-----------|------|-------------|
| `precipitation_amount` | mm | Total precipitation for the period |
| `air_temperature_max` | celsius | Maximum temperature during the period |
| `air_temperature_min` | celsius | Minimum temperature during the period |
| `symbol_code` | string | Weather symbol identifier (see Weather Icons below) |

The exact set of parameters depends on your configuration file (`jsonformat.json`). See [Configuration](#configuration) below.

## Weather Icons

The `symbol_code` in the `summary` section corresponds to weather icon filenames from the [MET Norway Weather Icon](https://github.com/metno/weathericons) set. For example:

- `clearsky_day` → clearsky_day.svg
- `partlycloudy_night` → partlycloudy_night.svg
- `rain` → rain.svg

Icons include variations for day, night, and polar conditions. No additional calculations are needed to select the correct day/night variant.

## Error Responses

### 400 Bad Request

Invalid or missing parameters.

```json
Missing value for lat
```

```json
lat must be in range -90 to 90
```

### 404 Not Found

The requested location is outside the coverage area.

```json
Outside of coverage area
```

### 200 OK with Error Message

When a point is within a grid's bounding area but too far from any grid cell, a 200 response is returned with an error in the GeoJSON body:

```json
{
  "type": "Feature",
  "geometry": {
    "type": "Point",
    "coordinates": [10.72, 59.93]
  },
  "properties": {
    "meta": {
      "updated_at": "2024-06-10T13:04:26Z",
      "error": "no data at the given location",
      "units": {}
    },
    "timeseries": []
  }
}
```

### 503 Service Unavailable

The service cannot retrieve forecast data (e.g., upstream service unavailable).

```json
service unavailable
```

## HTTP Headers

### Request Headers

**Accept-Encoding**

The API supports gzip compression. Include `Accept-Encoding: gzip` in your request to receive compressed responses (if enabled in configuration).

### Response Headers

| Header | Description |
|--------|-------------|
| `Content-Type` | `application/json; charset=utf-8` |
| `Content-Encoding` | `gzip` (if compression is enabled and requested) |
| `Last-Modified` | Current server time when response was generated |
| `Expires` | Cache expiration time (randomized to prevent thundering herd) |

Additional custom headers may be configured via `jsonformat.json`.

## Configuration

The jsonfrontend behavior is controlled by a configuration file (default: `jsonformat.json`). Key configuration options include:

- **Parameters mapping**: Maps internal parameter names to external/API names
- **Time periods**: Defines which time aggregation periods are exposed (`next_1_hours`, `next_6_hours`, etc.)
- **Units**: Unit definitions for each parameter
- **HTTP headers**: Custom HTTP response headers
- **Gzip compression**: Enable/disable gzip response compression
- **Altitude handling**: Whether to include altitude in responses
- **Forecast cutoff**: Whether to exclude past forecast times

See your Forti deployment's configuration file for the specific parameters available in your instance.

## Rate Limiting and Caching

Forti does not implement built-in rate limiting. The `Expires` header uses a randomized offset (if configured) to spread cache invalidation across clients and prevent cache stampedes.

It is recommended to:
- Respect the `Expires` header for caching
- Avoid excessive requests for the same location
- Use appropriate User-Agent headers to identify your application

## Legacy Frontends

**Note:** The `xmlfrontend` and `moxfrontend` services are legacy components and should not be used for new integrations. Use `jsonfrontend` for all new applications.

## Metrics

When metrics are enabled (`-metricsPort` flag), Prometheus metrics are exposed on the configured port (default: 9090) at `/metrics`.

## Further Reading

- [MET Norway Locationforecast 2.0 Documentation](https://api.met.no/weatherapi/locationforecast/2.0/documentation)
- [Locationforecast Data Model](https://api.met.no/doc/locationforecast/datamodel)
- [General Forecast JSON Format](https://api.met.no/doc/ForecastJSON)
- [Weather Icons](https://github.com/metno/weathericons)
