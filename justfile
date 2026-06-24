build-docker *modules="correctedforecaster healthz jsonfrontend moxfrontend rawdataforecaster xmlfrontend":
    set -e; for module in {{ modules }}; do \
        docker build -t forti_$module -f $module/build/package/Dockerfile .; \
    done
