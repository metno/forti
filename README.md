# Forti

## What is it?

Forti is a REST webservice that delivers weather and ocean forecast timeseries data for a specified lat/long coordinate. 

The data sources to the service are irregular grids of data, produced by batch jobs. These batch jobs use a wide range of input datasets, and run a set of post-processing algorithms to both improve the forecast quality and to add additional forecast parameters.

The application is built with Go, although with some small parts in C++.

The code for the batch jobs that produce the datasets are not included in this repository.

## Usage

The application is meant to be run as containers. Each component has an associated Dockerfile.

## Development

### Test

### Build

For all components, except fortiup, a docker image is built and pushed on every commit.
They will all get the same tag, corresponding to `$CI_PIPELINE_ID`, which identifies a single run of a gitlab ci pipeline.

For example, you may end up with the following images in the registry.

- fortiregistry.azurecr.io/xmlfrontend:337359
- fortiregistry.azurecr.io/rawdataforecaster:337359
- fortiregistry.azurecr.io/jsonfrontend:337359
- fortiregistry.azurecr.io/correctedforecaster:337359
- fortiregistry.azurecr.io/rawdataforecaster:337359
- fortiregistry.azurecr.io/healthz:337359

In this case the docker tag, 337359, corresponds to the gitlab's pipeline id when building the project.
The pipeline id may be found by looking at the correct build in the [pipelines](https://gitlab.met.no/team-punkt/forti/f2/-/pipelines) page.

## Run locally

[How to run f2 locally](run_locally.md)

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
C4Container
  title C4 Container Diagram - Forti

  Person_Ext(client, "API Client", "Consumes forecast data over REST")

  Boundary(forti, "Forti") {
    System(ingress, "Ingress", "Routes incoming requests")
    Container(dataloader, "Dataset Loader", "CLI", "Loads forecast datasets into the object store")
    System(objectstore, "Object Store", "Stores forecast datasets")
    Boundary(core, "Core") {
      Container(frontends, "Frontends", "Go", "Serve point forecast timeseries over REST")
      Container(correctedforecaster, "Correctedforecaster", "Go GRPC", "Post-process and serve forecast data")
      Container(rawdataforecaster, "Rawdataforecaster", "Go GRPC", "Serve forecast from cache or object store")
    }
    Boundary(monitoring, "Monitoring") {
      Container(healthz, "Healthz", "Go", "Runs integration tests and reports health over REST")
      System(prometheus, "Prometheus", "Collects metrics from Forti components")
    }
  }

  Rel(client, ingress, "REST")
  Rel_Down(ingress, frontends, "http")
  Rel_Right(ingress, healthz, "http")
  Rel_Right(ingress, prometheus, "http")
  Rel_Left(healthz, ingress, "rest")
  Rel(frontends, correctedforecaster, "grpc")
  Rel(correctedforecaster, rawdataforecaster, "grpc")
  Rel(correctedforecaster, objectstore, "Download topography")
  Rel(rawdataforecaster, objectstore, "Read forecast data")
  Rel(dataloader, objectstore, "Upload forecast data")
  Rel_Left(prometheus, frontends, "http")
  Rel_Left(prometheus, correctedforecaster, "http")
  Rel_Left(prometheus, rawdataforecaster, "http")
```
