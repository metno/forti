# rawdataforecaster

Serves raw (uncorrected) forecast data over gRPC from a blob store (Azure, S3, or local `file://`) in the [forti-internalformat](https://github.com/metno/forti-internalformat) format.

## Usage

See `rawdataforecaster --help` for command-line options.

## Configuration

The configuration file is JSON. Example:

```json
{
  "source": {
    "bucket": "file:///data/forecast"
  },
  "areas": ["meps"],
  "loader": {
    "type": "blob"
  },
  "maximum_gridpoint_distance": 50000
}
```

### Configuration Options

#### `source`

Defines where forecast data is stored.

- **`source.bucket`** (string, required): Bucket URL where forecast datasets are stored. Supported schemes:
  - `file:///absolute/path` — Local filesystem
  - `azblob://container-name` — Azure Blob Storage (requires Azure credentials)
  - `s3://bucket-name` — Amazon S3 (requires AWS credentials)

#### `areas`

- **`areas`** (array of strings, required): List of forecast area identifiers to load. Each area corresponds to a subdirectory in the bucket containing forecast datasets in forti-internalformat. Example values:
  - `"meps"` — MetCoOp Ensemble Prediction System (Nordic region)
  - `"tanzania"` — Some forecast for Tanzania

#### `loader`

Defines how forecast data is loaded and cached.

- **`loader.type`** (string, required): Loading strategy. See [Loader strategies](#loader-strategies) below.
- **`loader.configuration`** (object, optional): Strategy-specific options (see below).

#### `maximum_gridpoint_distance`

- **`maximum_gridpoint_distance`** (integer, optional): Maximum distance in meters from a requested point to the nearest grid cell. Points farther away return a "point too far away" error. If omitted, will return the nearest point, regardless of distance.

## Loader strategies

The `loader.type` config key selects how forecast data is held in memory:

| Type | Behaviour | Use When |
|---|---|---|
| `memory` | Downloads the full grid blob at startup into CGo-allocated memory (bypassing GC pressure). Zero I/O at query time. Capped by `max_size_gib`. | Latency matters and the dataset fits in RAM |
| `blob` | Downloads nothing upfront; issues a byte-range read per query (2 s timeout). | Dataset is too large for RAM, or startup time matters |

Use `memory` when latency matters and the dataset fits in RAM; use `blob` otherwise.

### Memory Loader Configuration

When using `loader.type: "memory"`, you can configure:

```json
{
  "loader": {
    "type": "memory",
    "configuration": {
      "max_size_gib": 4
    }
  }
}
```

- **`max_size_gib`** (number, optional): Maximum size in GiB for each loaded grid. Grids larger than this are rejected at startup.

### Blob Loader Configuration

The `blob` loader has no additional configuration options.

## Native dependencies

The geographic index uses [s2geometry](https://github.com/google/s2geometry) and [PROJ](https://proj.org/) via CGo. These are pre-installed in the devcontainer; building outside it requires both libraries.
