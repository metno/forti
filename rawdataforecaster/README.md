# Simple Forecaster

Serves unmodified data from a forecast. The term 'simple' refers to the data not having been processed by forti.

## Serving data

```mermaid
graph LR;
server --> forecast;
forecast --> fortidb;
fortidb --> index;
fortidb --> values;
```

Four components are mainly involved in serving data: 

### server

Handles incoming grpc requests, including protobuf serialization.

### forecast

Determines what is the correct group (and version) to serve data from. Forwards requests to the relevant `datagroup`.

### fortidb

Each object of type `fortidb.Dataset` serves data for a single group/verision. They maintain a list of grids (aka "hashes") for its group.

Handles requests for a given latitute/longitude pair. For each hash, lookup the correct index from `index`, and find relevant data from `values`.

### index

Handles lookup from latitude/longitude to a grid index.

### values

A collection of all data having the same group. 

Provides a `Reader` interface, for looking up data with a given index. The index is provided by the `geo` component. There are several implementations of this interface.

## Loading data

`forecast` component contains a function, `Forecast.update`, that is called periodically in a goroutine. It checks a blob store for updates, and loads data if needed, by calling `datagroup.Download`.

## Other modules

### internal.health

Provides grpc healthcheck, meant for kubernetes readiness probe.

### pointdata

Defines internal data format. Its placement reflects that several modules in various places in the hierarchy needs access to this.
