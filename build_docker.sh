#!/bin/bash

set -euo pipefail

echo $@

PREFIX="$1"
NAME=$(dirname $2)
VERSION=$(basename $2)

IMAGE="${PREFIX}/${NAME}:${VERSION}"

echo $IMAGE

docker build -t "$IMAGE" -f "$NAME/build/package/Dockerfile" .

docker push "$IMAGE"
