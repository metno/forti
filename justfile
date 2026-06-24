build-docker *modules="correctedforecaster healthz jsonfrontend moxfrontend rawdataforecaster xmlfrontend":
    for module in {{ modules }}; do \
        docker build -t forti_$module -f $module/build/package/Dockerfile .; \
    done
