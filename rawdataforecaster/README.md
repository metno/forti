# rawdataforecaster

Serves raw (uncorrected) forecast data over gRPC from a blob store (Azure, S3, or local `file://`) in the [forti-internalformat](https://github.com/metno/forti-internalformat) format.

## Loader strategies

The `loader.type` config key selects how forecast data is held in memory:

| Type | Behaviour |
|---|---|
| `memory` | Downloads the full grid blob at startup into CGo-allocated memory (bypassing GC pressure). Zero I/O at query time. Capped by `max_size_gib`. |
| `blob` | Downloads nothing upfront; issues a byte-range read per query (2 s timeout). |

Use `memory` when latency matters and the dataset fits in RAM; use `blob` otherwise.

## Native dependencies

The geographic index uses [s2geometry](https://github.com/google/s2geometry) and [PROJ](https://proj.org/) via CGo. These are pre-installed in the devcontainer; building outside it requires both libraries.
