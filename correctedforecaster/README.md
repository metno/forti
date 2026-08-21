# correctedforecaster

gRPC service that applies post-processing corrections to forecast data. It fetches raw forecasts from an upstream gRPC server (typically `rawdataforecaster`), applies corrections, and exposes the corrected data via the same gRPC interface.

## What it does

correctedforecaster applies location-specific corrections to forecast data to improve accuracy:

- **Adiabatic temperature correction**: Adjusts all temperature parameters (including apparent temperature) based on elevation differences between the forecast grid and actual ground elevation using topography data. Uses a standard adiabatic lapse rate of 0.6°C per 100m.
- **Weather symbol adjustment**: Updates precipitation phase in weather symbols (rain/sleet/snow) based on the corrected temperatures
- **Additional post-processing**: Other model output corrections as implemented

## Usage

See `correctedforecaster --help` for command-line options.

The service listens on port `5051` (configurable) for incoming gRPC requests.

## Topography Data

correctedforecaster requires topography (elevation) data files to perform temperature corrections. These files must be available before starting the service.

### Data Sources

Topography data can be sourced in two ways:

1. **Download from bucket** (recommended for production):
   ```bash
   correctedforecaster -upstream rawdataforecaster:5052 \
     -download-from azblob://topography-container \
     -workdir /data/topography
   ```
   
   Downloads topography files from the specified bucket to `-workdir` at startup.

2. **Use local files** (for development):
   ```bash
   correctedforecaster -upstream rawdataforecaster:5052 \
     -workdir /data/topography
   ```
   
   Uses topography files already present in `-workdir`. No download occurs.

### Bucket URL Formats

Supported bucket URL schemes (same as `rawdataforecaster`):
- `file:///absolute/path` — Local filesystem
- `azblob://container-name` — Azure Blob Storage (requires Azure credentials)
- `s3://bucket-name` — Amazon S3 (requires AWS credentials)

### Required Files

The topography data directory must contain elevation data files that cover the geographic regions served by your deployment. For effective temperature correction, topography data should have higher spatial resolution than the forecast model grid.

Topography files can be in any raster format readable by GDAL (GeoTIFF, NetCDF, HDF5, etc.). The files must contain elevation data with proper geospatial projection information.

## Architecture

```
client → jsonfrontend → correctedforecaster → rawdataforecaster → blob storage
                             ↓
                       topography data
```

- **Upstream**: `rawdataforecaster` (or another gRPC forecaster)
- **Downstream**: Frontend services like `jsonfrontend`
- **Data dependency**: Topography files (loaded at startup)

## Configuration

correctedforecaster does not use a configuration file. All options are specified via command-line flags (see `--help`).

## No Configuration File

Unlike other Forti components, correctedforecaster does not use a JSON configuration file. All behavior is controlled by command-line flags and the topography data files.
