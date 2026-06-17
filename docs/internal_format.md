# Forti's internal data format

This document is intended for developers who need to read or write Forti’s internal data format directly.

Forti uses an internal data format for storing data. This format is used both in the blob storage and by `rawdataforecaster`. 

## Overview

Data is split into areas and versions.
An area refers to to a limited geographic area, typically the domain of a single forecast model. 
Version numbers are used to identify which forecast for an area is the newest.
Highest version number is newest.

A single rawdataforecaster instance can serve data from several areas, but only one area for a single request. 
This is done by selecting the area with a grid point which is closest to the requested location.
It will only ever serve the latest version for each area, as determined by that area's version number.

The data in the single area is expressed as one or more lists of latitude/longitude values with accompanying data for several parameters.
If there are more than one list of latitudes and longitudes, they are expected to cover the same area, but with different values for latitude and longitude.
This allows some parameters to have a different resolution than others, even if the cover the same area.

All data for each area/version is placed in a uniquely-named subfolder under `/<area>/<version>/` in the blob storage. 
The structure of this is described below.


## Object store layout

Under a single area/version in the blob storage, the following layout is expected:

* complete.json
* sub-folders, containing the following objects:
  * meta.json
  * data
  * longitude
  * latitude

### complete.json

This contains metadata about the area/version itself. 
Its format is described in `DatasetMeta` in [the code](../fortiup/pkg/fortiblob/collector.go).

When uploading data to the object store, this is supposed to be the last file uploaded, as its existence will trigger an update on `rawdataforecaster`.

### Sub-folders

Different data for the same geographic area can have different resolutions.
This will be expressed as different values for longitude and latitudes.
For each of these resolutions, a subfolder is made.

The name of this sub-folder is expected to be unique for each set of lat/lon lists.
For example, the name can be equal to the md5 sum of the concatenated latitude and longitude lists.

Four files are expected to exist here:

* meta.json
* longitude
* latitude
* data

#### meta.json

This describes the meaning of the forecast values. 
The data structure is described in `MetaCollection` in [the code](../fortiup/pkg/fortiblob/collector.go). An example of such data can be found in [the same folder](../fortiup/pkg/fortiblob/collector_test.go)

#### longitude and latitude

In the blob store, latitudes and longitudes exist in two separate files. 
These define the lat/lon of each forecast point in the data, and each latitude/longitude is binary encoded as a `float32` in the files.
The ordering of the values are not important, as long as the same ordering is used in the logintude, latitude and data files.

#### data

Data contains the actual values for the forecast. 
It consists of a series of little-endian encoded `ìnt16` values, and their meaning is defined in meta.json.

To look up data for a specific location, you need two things:
* An index from the latitude and longitude arrays.
* The length of the relevant data - this is the metadata's number_of_points value.

Multiply the two values to get the starting index.
You can then use the metadata to interpret the relavant values.
[The example](../fortiup/pkg/fortiblob/collector_test.go) shows how to do the interpretation.
