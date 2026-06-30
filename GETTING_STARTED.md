# Getting Started

This guide walks you through running Forti locally using Docker Compose.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Compose support
- [just](https://just.systems/) (optional, but convenient)

## 1. Prepare forecast data

Forti reads forecast data from a local directory in [its own binary format](https://github.com/metno/forti-internalformat). You need to populate `data/forecast/` (relative to this repository root) before starting the stack — this is where Docker Compose expects data by default.

### Using forti-prep (recommended)

[forti-prep](https://github.com/metno/forti-prep) is a tool that converts NetCDF forecast files into the format Forti expects. Requires [uv](https://docs.astral.sh/uv/).

```bash
git clone https://github.com/metno/forti-prep
cd forti-prep
uv sync
uv run forti-prep \
  --config sample_config.json \
  --output-dir /path/to/forti/data/forecast \
  --version $(date +%s) \
  your-forecast.nc
```

See the [forti-prep README](https://github.com/metno/forti-prep) for details on writing a config file for your specific NetCDF input.

### Writing your own loader

If you have a different data source, you can produce the data directly in Forti's internal format. See [forti-internalformat](https://github.com/metno/forti-internalformat) for a full description of the format and Go code you can use to write it.

## 2. Start the stack

From the repository root, using just:

```bash
just run-docker
```

Or directly with Docker Compose:

```bash
cd deploy
docker compose up --build
```

This starts `rawdataforecaster` and `jsonfrontend`. The first run will build the Docker images, which may take a few minutes.

## 3. Verify

```bash
curl 'http://localhost:8080/?lat=59&lon=11'
```

You should get a JSON forecast response.

## Next steps

See [`deploy/README.md`](deploy/README.md) for configuration options, including how to enable `correctedforecaster` and how to override data paths.
