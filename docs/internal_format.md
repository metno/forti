# Forti's internal data format

Forti uses an internal data format for storing data. This format is used both in the blob storage and by `rawdataforecaster`. It also dictates parts of the contents of the grpc protocol.
Each area/version consists of one or several separate forecasts grouped by their grid resolution.

## Data for each area/version

Each area/version has one or several grids (strictly speaking, it is not a grid, but a collection of latitude/longitude coordinates within a particular area). Each grid is supposed to have the same geographic extent, but differs in what resolution the grid points in each grid has.

The data for each grid resolution is described in a uniquely-named subfolder under `/<area>/<version>/` in the blob storage. The data for each resolution is described below.

Each combination of `area` and `version` has accompanying metadata. These data are described in `DatasetMeta` in [the code](../upload/pkg/fortiblob/collector.go). On the blob store, these data are stored in a file called `complete.json`.


## Format for each grid resolution

The format consists of four pieces of data:

* forecast values
* metadata
* latitudes
* longitudes

### Forecast values

The forecast values are the actual values for the forecast. The data is organized so that all values for each location is stored together, one location after the other. Each value in the data is merely a little-endian encoded `int16`, and the meaning of each one is defined in the metadata.

On the blob store, these data are stored in a file called `data`.

### Metadata

The metadata describes the meaning of the forecast values. The data structure is described in `MetaCollection` in [the code](../upload/pkg/fortiblob/collector.go). There is also [an example](../upload/pkg/fortiblob/collector_test.go) available for how to interpret raw data.

In the blob store, this is json-encoded in a file called `meta.json`.

### Latitudes and longitudes

In the blob store, latitudes and longitudes exist in two separate files. These define the lat/lon of each forecast point in the data, and each latitude/longitude is binary encoded as a `float32` in the files.

The index of each lat/lon pair matches the index of the forecast in the `data` file. So, if you want to look up the data for latitude/longitude index `x`, that data starts in the data file at index `(x * meta.LocationCount)`. Of course, if you make a lookup into a file, you must multiply this by 2 to account for int16 taking up two bytes.
