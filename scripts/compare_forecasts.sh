#!/bin/bash

set -euo pipefail

TYPE=$1
SERVER_A=$2
SERVER_B=$3

LOCATIONS="lat=78.2167&lon=15.6333 lat=59&lon=11 lat=-68.1594&lon=-127.2076" 

SERVICES="complete compact classic.xml"
SERVICES="complete classic.xml"
# SERVICES="classic.xml"

for service in $SERVICES; do    
    for location in $LOCATIONS; do
        url="https://${SERVER_A}.forti.met.no/api/${TYPE}/v2/${service}?${location}"
        echo "$url"
        curl -fso "/tmp/${SERVER_A}" "$url" || continue

        url="https://${SERVER_B}.forti.met.no/api/${TYPE}/v2/${service}?${location}"
        echo "$url"
        curl -fso "/tmp/${SERVER_B}" "$url"

        if file "/tmp/${SERVER_A}" | grep -i json; then
            diff <(jq -S . "/tmp/${SERVER_A}") <(jq -S . "/tmp/${SERVER_B}") || true
        else
            diff "/tmp/${SERVER_A}" "/tmp/${SERVER_B}" || true
        fi
    done
done
