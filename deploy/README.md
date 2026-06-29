# Deploy

This folder contains configuration for running Forti locally using Docker Compose. It brings up the core services — `rawdataforecaster`, `correctedforecaster`, and `jsonfrontend` — using local forecast data.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Compose support
- Local forecast data (see [Preparing data](#preparing-data) below)

## Preparing data

Forecast data can be produced using [forti-prep](https://github.com/metno/forti-prep), a companion tool that downloads and post-processes the input datasets into the format expected by Forti.

By default, the Compose setup expects forecast data to be available at `../data/forecast` (relative to this folder), which corresponds to `data/forecast/` in the repository root. You can override this with the `FORECAST_DATA_PATH` environment variable.

## Services

| Service | Description | Default port |
|---|---|---|
| `rawdataforecaster` | Serves forecast data over gRPC from a local directory | `5052` |
| `correctedforecaster` | Post-processes forecast data and re-exposes it over gRPC | — |
| `jsonfrontend` | REST API that serves point forecast timeseries as JSON | `8080` |

`correctedforecaster` is optional and must be enabled via a Docker Compose [profile](#profiles).

## Usage

Build and start the base services:

```bash
docker compose up --build
```

Test that the API is responding:

```bash
curl 'http://localhost:8080/?lat=59&lon=11'
```

### Enabling correctedforecaster

`correctedforecaster` is disabled by default. To enable it, set both variables in your `.env` file:

```bash
COMPOSE_PROFILES=corrected
JSONFRONTEND_UPSTREAM=correctedforecaster:5051
```

Both need to change together: `COMPOSE_PROFILES` starts the container, and `JSONFRONTEND_UPSTREAM` points `jsonfrontend` at it. Leave `COMPOSE_PROFILES` empty (or unset) to run without correction.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `FORECAST_DATA_PATH` | `../data/forecast` | Path to the local forecast data directory |
| `TOPOGRAPHY_DATA_PATH` | `../data/topography` | Path to the local topography data directory (used by `correctedforecaster`) |
| `JSONFRONTEND_UPSTREAM` | `rawdataforecaster:5052` | gRPC upstream address for `jsonfrontend` |

These can be set in a `.env` file in this directory (`.env` is gitignored). Copy `.env.example` as a starting point:

```bash
cp .env.example .env
```

## Configuration

`rawdataforecaster.config.json` configures how `rawdataforecaster` reads forecast data:

```json
{
  "source": {
    "bucket": "file:///data/forecast"
  },
  "areas": ["meps"],
  "loader": {
    "type": "blob"
  }
}
```

- **`source.bucket`** — points to the mounted forecast data directory inside the container.
- **`areas`** — list of forecast areas to load (e.g. `meps`).
- **`loader.type`** — `"blob"` streams data on demand rather than loading everything into memory. Keep this as `"blob"` unless you have a specific reason to change it.
