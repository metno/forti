# Running locally

## Rawdataforecaster

### Install dependencies

There are a number of dependencies for building and running `rawdataforecaster`. See the application's [Dockerfile](rawdataforecaster/build/package/Dockerfile) for what to install and how to install it.

### Setup environment

If you have access to forti's blob store, you can read data directly from there. There is a [script](scripts/setup_credentials.sh) to setup the required environment variables for accessing the storage. Run it like this:

```bash
. scripts/setup_credentials.sh
```

This will grant temporary access to the staging blob storage. You will need to rerun this script every day when you want to work with data from the blob storage.

### Run the application

Several workable config files have been added under `rawdataforecaster/cmd/rawdataforecaster/`. Use one of them (or create your own) and run the application like this:

```bash
cd rawdataforecaster/cmd/rawdataforecaster/
go run main.go -config <CONFIG>.json
```

Note that unless you set the config file's `loader.type` to "blob", your application may use huge amounts of memory.

### Test that it works

```bash
cd rawdataforecaster/cmd/lookup
go run main.go
```


## Correctedforecaster

Note that it is not neccessary to run the `correctedforecaster` to test the various frontends. Each of those can connect directly to the `rawdataforecaster`.

### Install dependencies

`correctedforecaster` requires `libgdal`. Install like this:

```bash
sudo apt install libgdal-dev
```

### Download data

Before you can begin, you need local access to topography data. Create a folder somewhere on your computer, and download all (or some) files from the `topography` bucket in the `modelstaging` cluster into there.

One way to download the data is to setup azure credentials, and then add `-download-from azblob://topography` as an additional option the first time you run the `correctedforecaster`. Note that this will download several gigabytes of data.

You can set up credentials like this:

```bash
export AZURE_STORAGE_ACCOUNT=modelstaging
export AZURE_STORAGE_SAS_TOKEN=$(az storage container generate-sas \
    -n topography \
    --subscription FortiSubscription \
    --account-name "$AZURE_STORAGE_ACCOUNT" \
    --https-only \
    --permissions lr \
    --expiry "$(date -d "1 day" +"%Y-%m-%dT00:00Z")" \
    -otsv \
)
```

### Run the application

```bash
cd correctedforecaster/cmd/correctedforecaster/
go run main.go -workdir <DATADIR>
```

### Test that it works

```bash
cd rawdataforecaster/cmd/lookup
go run main.go -address localhost:5051
```

## Frontends

All frontends (`jsonfrontend`, `xmlfrontend`, etc) work the same way. Here we use the `jsonfrontend` as an example.

```bash
cd jsonfrontend/cmd/jsonfrontend/
go run main.go
``` 

This will connect to the `correctedforecaster`. It is possible to bypass that one, and go straight to the `radataforecaster`. To do that, add an extra option when calling the program:

```bash
cd jsonfrontend/cmd/jsonfrontend/
go run main.go -upstream localhost:5052
``` 

### Test that it works

```bash
curl 'http://localhost:8080/?lat=59&lon=11'
``` 


