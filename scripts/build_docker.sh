#!/bin/bash

# This is a simple script to verify that all docker builds work
# It must be run from the repository root

set -euo pipefail

MODULES=$(find . -name Dockerfile | cut -d/ -f2)

for module in $MODULES; do 
	echo "Building $module"
	docker build -t f2_$module -f $module/build/package/Dockerfile  .
done

