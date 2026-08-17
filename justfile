image_repo := "forti"
image_tag := "latest"

build-docker *modules="correctedforecaster healthz jsonfrontend moxfrontend rawdataforecaster xmlfrontend":
    set -e; for module in {{ modules }}; do \
        docker build -t {{image_repo}}$module:{{image_tag}} -f $module/build/package/Dockerfile .; \
    done

run-docker:
    cd deploy && IMAGE_REPO={{image_repo}} IMAGE_TAG={{image_tag}} docker compose up
