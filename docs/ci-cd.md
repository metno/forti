# CI/CD

## Container images

A Docker image is built and pushed to the registry on every commit. All images from a single workflow run share the same tag, corresponding to the GitHub Actions `run_id`.

For example:

- `fortiregistry.azurecr.io/xmlfrontend:337359`
- `fortiregistry.azurecr.io/moxfrontend:337359`
- `fortiregistry.azurecr.io/rawdataforecaster:337359`
- `fortiregistry.azurecr.io/jsonfrontend:337359`
- `fortiregistry.azurecr.io/correctedforecaster:337359`
- `fortiregistry.azurecr.io/healthz:337359`

The tag (`337359` above) is the GitHub Actions run ID for the build. You can find the run ID for a specific commit on the [Actions](https://github.com/metno/forti/actions) page.
