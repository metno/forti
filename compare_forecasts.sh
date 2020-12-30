#!/bin/bash

set -euo pipefail

SERVER_A=$1
SERVER_B=$2

POSITION="lat=60&lon=10"

curl "https://${SERVER_A}.forti.met.no/api/forecast/v2/complete?$POSITION" | jq . > "/tmp/${SERVER_A}.json"
curl "https://${SERVER_B}.forti.met.no/api/forecast/v2/complete?$POSITION" | jq . > "/tmp/${SERVER_B}.json"

diff "/tmp/${SERVER_A}.json" "/tmp/${SERVER_B}.json"