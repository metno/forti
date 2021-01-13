# Simple Forecaster

Serves unmodified data from a forecast. The term 'simple' refers to the data not having been processed by forti.

## Serving data

```mermaid
graph LR
server --> forecast
forecast -> datagroup
datagroup --> geo
datagroup --> simpledatagroup
```

Four components are mainly involved in serving data: 

### server

Handles incoming grpc requests, including protobuf serialization.

### forecast

Determines what is the correct group (and version) to serve data from. Forwards requests to the relevant `datagroup`.

### datagroup

Each object of type `datagroup.Dataset` serves data for a single group/verision. They maintain a list of grids (aka "hashes") for its group.

Handles requests for a given latitute/longitude pair. For each hash, lookup the correct index, and find relevant data.

### geo

Handles lookup from latitude/longitude to a grid index.

### simpledatagroup

A collection of all data having the same group. 

Provides a `Reader` interface, for looking up data with a given index. The index is provided by the geo component. There are several implementations of this interface.

