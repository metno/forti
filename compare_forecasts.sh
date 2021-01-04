#!/bin/bash

set -euo pipefail

SERVER_A=$1
SERVER_B=$2

URL_PATH="/api/forecast/v2/complete"
POSITION="lat=-89&lon=0"

curl "https://${SERVER_A}.forti.met.no${URL_PATH}?${POSITION}" | jq . > "/tmp/${SERVER_A}"
curl "https://${SERVER_B}.forti.met.no${URL_PATH}?${POSITION}" | jq . > "/tmp/${SERVER_B}"

diff "/tmp/${SERVER_A}" "/tmp/${SERVER_B}"
