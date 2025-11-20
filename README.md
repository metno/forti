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
flowchart LR
 subgraph Met["Met"]
        api_met_no["api.met.no (EKS)<br>[Software-system]"]
        ppi["ecflow ppi jobs (INT)<br>[Software-system]"]
        grafana["Team Punkt's grafana server (INT)<br>[Software-system]"]
  end
 subgraph Core["Core"]
        Frontends["<b>Frontends</b> (INT)<br>[Container: Go web servers]<br>Several systems. Serve point forecast timeseries over REST in geojson, xml or other formats.<br>Each instance differs in output format only."]
        Correctedforecaster["<b>Correctedforecaster</b> (INT)<br>[Container: Go GRPC server]<br>Request forecast data, adjust and serve."]
        Rawdataforecaster["<b>Rawdataforecaster</b> (INT)<br>[Container: Go GRPC server]<br>Serve forecast timeseries from memory cache or blob storage."]
  end
 subgraph Kubernetes["Kubernetes"]
        Ingress["Ingress controller (EKS)"]
        Core
        Healthz["Healthz (INT)<br>[Container: Go web server]<br>Periodically run integration tests and deliver status over REST."]
        Prometheus["Prometheus (INT)<br>[Software system]<br>Provides insights into performance and other statistics about each forti component. Prometheus web server is password-protected"]
  end
 subgraph Azure["Azure"]
        Azureblob["Azure Blob Storage (SAS)<br>[Software-system]"]
        Kubernetes
  end
 subgraph Forti["Forti"]
        Azure
        fortiup["fortiup (INT)<br>[Container: Go command-line program]<br>Upload Forti netcdf dataset from filesystem."]
  end

yr["YR (EKP)<br>[Software-system]"]
pingdom["Pingdom (EXT)<br>[Software-system]"]

    Correctedforecaster -->|"Download topography<br>(level 2) "| Azureblob
    Ingress -->|"http<br>(level 2) "| Frontends
    Ingress -->|"https<br>(level 2) "| Healthz
    Ingress -->|"http<br>(level 2)  "| Prometheus
    Frontends -->|"grpc req/reply<br>(level 2) "| Correctedforecaster
    Correctedforecaster -->|"grpc req/reply<br>(level 2) "| Rawdataforecaster
    Healthz -->|"rest<br>(level 2) "| Ingress
    Prometheus -->|"http<br>(level 2) "| Frontends & Correctedforecaster & Rawdataforecaster
    Rawdataforecaster -->|"Read forecast data<br>(level 2) "| Azureblob
    fortiup -->|"Upload forecast data<br>(level 2) "| Azureblob
    yr -->|"rest<br>(level 0) "| Ingress
    pingdom -->|"https<br>(level 0) "| Ingress
    ppi -->|"Calls<br>(level 2) "| fortiup
    grafana -->|"https<br>(level 2) "| Ingress
    api_met_no -->|"rest<br>(level 2) "| Ingress
```
