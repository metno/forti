#!/bin/bash

set -euo pipefail

SERVER_A=$1
SERVER_B=$2

URL_PATH="/weatherapi/locationforecast/1.9"
POSITION="lat=60&lon=10"

curl "https://${SERVER_A}.forti.met.no${URL_PATH}?${POSITION}"  > "/tmp/${SERVER_A}"
curl "https://${SERVER_B}.forti.met.no${URL_PATH}?${POSITION}"  > "/tmp/${SERVER_B}"

diff "/tmp/${SERVER_A}" "/tmp/${SERVER_B}"
