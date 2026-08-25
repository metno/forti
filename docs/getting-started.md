# Getting Started

This guide walks you through running Forti locally with `go run`.

For a Docker Compose-based setup, see [forti-deploy](https://github.com/metno/forti-deploy).

## Prerequisites

- Go 1.26 or later
- Native dependencies for `rawdataforecaster` and `correctedforecaster` (s2geometry, PROJ) — the devcontainer (`.devcontainer/`) has everything set up

## 1. Prepare forecast data

Forti reads forecast data from a local directory in [its own binary format](https://github.com/metno/forti-internalformat). You need to populate a forecast data directory before starting the services.

### Using forti-prep (recommended)

[forti-prep](https://github.com/metno/forti-prep) is a tool that converts NetCDF forecast files into the format Forti expects. Requires [uv](https://docs.astral.sh/uv/).

```bash
git clone https://github.com/metno/forti-prep
cd forti-prep
uv sync
uv run forti-prep \
  --config sample_config.json \
  --output-dir /path/to/your/forecast/data \
  --version $(date +%s) \
  your-forecast.nc
```

See the [forti-prep README](https://github.com/metno/forti-prep) for details on writing a config file for your specific NetCDF input.

### Writing your own loader

If you have a different data source, you can produce the data directly in Forti's internal format. See [forti-internalformat](https://github.com/metno/forti-internalformat) for a full description of the format and Go code you can use to write it.

## 2. Start the services

Create a configuration file for `rawdataforecaster` that points to your forecast data directory. Set `source.bucket` to a `file://` URL:

```json
{
  "source": {
    "bucket": "file:///path/to/your/forecast/data"
  }
}
```

Start the services in separate terminals:

```bash
# Terminal 1 – rawdataforecaster (gRPC on :5052)
go run ./rawdataforecaster/cmd/rawdataforecaster -config your-config.json

# Terminal 2 – jsonfrontend (HTTP on :8080)
go run ./jsonfrontend/cmd/jsonfrontend -upstream localhost:5052
```

## 3. Verify

```bash
curl 'http://localhost:8080/?lat=59&lon=11'
```

You should get a JSON forecast response.

## Next steps

For detailed component configuration:
- [jsonfrontend configuration](../jsonfrontend/README.md)
- [rawdataforecaster configuration](../rawdataforecaster/README.md)
- [correctedforecaster configuration](../correctedforecaster/README.md)

To enable `correctedforecaster`, see the [correctedforecaster README](../correctedforecaster/README.md) for setup instructions.
