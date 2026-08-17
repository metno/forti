# Download Topography

A Go tool for downloading Copernicus GLO-30 DEM (Digital Elevation Model) tiles from AWS S3. The tiles are 30-meter resolution elevation data useful for topographic correction in weather forecasting.

## Building

From the repository root:

```bash
go build ./tools/download-topography/cmd/download-topography
```

Or run directly:

```bash
go run ./tools/download-topography/cmd/download-topography [flags]
```

## Prerequisites

None! The tool downloads tiles via HTTPS from the public Copernicus DEM S3 bucket.

## Usage

### Predefined Regions

Download using a predefined region:

```bash
# Kenya (~100 tiles, ~10GB)
./download-topography --region kenya --output ../data/topography

# Malawi (~36 tiles, ~3-4GB)
./download-topography --region malawi --output ../data/topography

# All of Africa (~2000 tiles, ~200GB - not recommended unless necessary)
./download-topography --region africa --output ../data/topography
```

Available regions:
- `africa` — entire continent
- `kenya`
- `malawi`
- `nigeria`
- `ethiopia`
- `tanzania`
- `south-africa`

### Custom Bounding Box

Specify any rectangular area using lat/lon coordinates:

```bash
./download-topography \
  --lat-min -12 \
  --lat-max 5 \
  --lon-min 29 \
  --lon-max 42 \
  --output ../data/topography
```

### Additional Options

```bash
# Dry run to preview what will be downloaded
./download-topography --region kenya --output ../data/topography --dry-run

# Use more parallel workers (default: 4)
./download-topography --region kenya --output ../data/topography --workers 8

# See all options
./download-topography --help
```

## Output

All `.tif` files are downloaded directly into the output directory:

```
data/topography/
├── Copernicus_DSM_COG_10_S01_00_E033_00_DEM.tif
├── Copernicus_DSM_COG_10_S01_00_E034_00_DEM.tif
├── Copernicus_DSM_COG_10_S01_00_E035_00_DEM.tif
└── ...
```

Each file is a Cloud-Optimized GeoTIFF containing the 30-meter resolution elevation data for that 1°×1° tile.

## Data Source

- **Source**: Copernicus DEM GLO-30
- **Resolution**: 30 meters
- **Format**: GeoTIFF (Cloud-Optimized)
- **Coverage**: Global
- **S3 Bucket**: `s3://copernicus-dem-30m/` (public, no credentials required)
- **License**: Free and open

## Notes

- Ocean-only tiles are not available and will be skipped with a message
- Downloads are parallelized for efficiency (default 4 workers)
- Each 1°×1° tile is approximately 100MB
