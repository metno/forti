# Forti
## What is it?
Forti is a REST webservice that delivers weather and ocean forecast timeseries data for a specified lat/long coordinate. 

The data sources to the service are irregular grids of data, produced by batch jobs. These batch jobs use a wide range of input datasets, and run a set of post-processing algorithms to both improve the forecast quality and to add additional forecast parameters.

The application is built with Go, although with some small parts in C++.

The code for the batch jobs that produce the datasets are not included in this repository.

## Usage
The application is meant to be run as containers. Each component has an associated Dockerfile.

## Run locally
TODO: Describe setting up local development version of Forti.


## Architecture
The application is made up of several binaries working together.

- jsonfrontend and xmlfrontend: serves the REST interface.
- healthz: monitors the overall health of the application.
- rawdataforecaster: delivers forecast through a GRPC interface, from either a in-memory cache or from a blob storage.
- correctedforecaster: collects data from the grpc interface, does some post-processing on the forecast and delivers the forecast through the same grpc interface.
- Azure blob storage: Object store containing the latest version of all the forecast data.
- Ecflow: Workflow manager specifying how and when to produce the forecast datasets.
- PPI: Compute, job scheduling and storage system for producing the forecast datasets. Storage system contains most of source data needed to produce the datasets.

### C4 container diagram
```mermaid
graph TD;
   YR["YR<br>[Software-system]"];
   api_met_no["api.met.no<br>[Software-system]"];

   YR-->|REST|JSONFrontend;
   api_met_no-->|REST|JSONFrontend;
   HealthDashboard-->|REST|Healthz;
   subgraph Forti
      JSONFrontend["JSONFrontend<br>[Container: Go web server]<br>Serve point forecast timeseries over REST in geojson."];
      Correctedforecaster["Correctedforecaster<br>[Container: Go GRPC server]<br>Request forecast data, adjust and serve."];
      Rawdataforecaster["Rawdataforecaster<br>[Container: Go GRPC server]<br>Serve forecast timeseries from memory cache or blob storage."];
      Azureblob["Azure Blob Storage<br>[Software-system]"];
      JSONFrontend-->|GRPC|Correctedforecaster;
      Correctedforecaster-->|GRPC|Rawdataforecaster;
      Rawdataforecaster-->|Read data|Azureblob;
      Healthz["Healthz<br>[Container: Go web server]<br>Periodically run integration tests and deliver status over REST."];
      Healthz-->|REST|JSONFrontend;
   end
   subgraph MET-infrastructure
      Ecflow["Ecflow<br>[Software-system]"];
      PPI["PPI<br>[Software-system]"];
      Fortiupload["Forti-upload<br>[Container: Ecflow PPI job]"];
      Fortiupload-->|Write data|Azureblob;
      Ecflow-->|Schedule jobs|PPI;
      PPI-->|Run job|Fortiupload;
   end
```

