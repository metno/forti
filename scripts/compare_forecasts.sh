#!/bin/bash

set -euo pipefail

SERVER_A=$1
SERVER_B=$2

URL_PATH="/api/forecast/v2/complete"

# POSITION="lat=60&lon=10"
# POSITION="lat=0&lon=10"
POSITION="lat=78.2167&lon=15.6333"

FORMAT="jq ."

curl "https://${SERVER_A}.forti.met.no${URL_PATH}?${POSITION}" | $FORMAT > "/tmp/${SERVER_A}"
curl "https://${SERVER_B}.forti.met.no${URL_PATH}?${POSITION}" | $FORMAT > "/tmp/${SERVER_B}"

diff "/tmp/${SERVER_A}" "/tmp/${SERVER_B}"
