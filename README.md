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

```plantuml
@startuml
!include  <C4/C4_Context.puml>
!include  <C4/C4_Container.puml>
!include <office/Users/user.puml>

Person_Ext(yr, "yr.no (EKP)", "yr.no including apps and web page")
System_Ext(pingdom, "Pingdom (EXT)", "Health checks")

System_Boundary(metno, "Met") {
  System_Ext(api_met_no, "api.met.no (EKS)")
  System_Ext(ecflow, "ecflow ppi jobs (INT)")
  System_Ext(grafana, "Grafana (INT)", "Team Punkt's grafana server")
}

System_Boundary(forti, "Forti") {
  Container(fortiup, "fortiup (INT)", "Go command-line program", "Upload Forti netcdf dataset from filesystem.")
  System_Boundary(azure, "Azure") {
    System(azureblob, "Azure Blob Storage (SAS)")
    System_Boundary(k8s, "Kubernetes") {
      System(ingress, "Ingress controller (EKS)")
      Container(healthz, "Healthz (INT)", "Go web server", "Periodically run integration tests and deliver status over REST.")
      System(prometheus, "Prometheus (INT)", "Provides insights into performance and other statistics about each forti component. Prometheus web server is password-protected.")
      System_Boundary(core, "Core") {
        Container(frontends, "Frontends (INT)", "Go web servers", "Several systems. Serve point forecast timeseries over REST in geojson, xml or other formats.<br>Each instance differs in output format only.")
        Container(correctedforecaster, "Correctedforecaster (INT)", "Go GRPC server", "Request forecast data, adjust and serve.")
        Container(rawdataforecaster, "Rawdataforecaster (INT)", "Go GRPC server", "Serve forecast timeseries from memory cache or blob storage.")
        Rel(frontends, correctedforecaster, "grpc req/reply (level 2)")
        Rel(correctedforecaster, rawdataforecaster, "grpc req/reply (level 2)")
      }
      Rel(ingress, frontends, "http (level 2)")
      Rel(ingress, healthz, "http (level 2)")
      Rel(ingress, prometheus, "http (level 2)")
      Rel(healthz, ingress, "rest (level 2)")
      Rel(correctedforecaster, azureblob, "Download topography (level 2)")
      Rel_R(prometheus, frontends, "http (level 2)")
      Rel_R(prometheus, correctedforecaster, "http (level 2)")
      Rel_R(prometheus, rawdataforecaster, "http (level 2)")
    }
    Rel_R(rawdataforecaster, azureblob, "Read forecast data (level 2)")
  }
  Rel(fortiup, azureblob, "Upload forecast data(level 2)")
}

Rel_D(grafana, ingress, "https (level 0)")
Rel_D(ecflow, fortiup, "calls program")

Rel_D(yr, ingress, "rest (level 0)")
Rel_D(api_met_no, ingress, "rest (level 0)")
Rel_D(pingdom, ingress, "https (level 0)")

@enduml```
