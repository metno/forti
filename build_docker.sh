#!/bin/bash

set -euo pipefail

PREFIX="$1"
NAME=$(dirname $2)
VERSION=$(basename $2)

IMAGE="${PREFIX}/${NAME}:${VERSION}"

echo build $IMAGE

docker build -t "$IMAGE" -f "$NAME/build/package/Dockerfile" .

# docker push "$IMAGE"
