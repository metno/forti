#!/bin/bash

set -euo pipefail

# GNU date (Linux) and BSD date (macOS) compatible
if date -d "1 day" +"%Y-%m-%dT00:00Z" >/dev/null 2>&1; then
    VALID_UNTIL=$(date -d "1 day" +"%Y-%m-%dT00:00Z")
else
    VALID_UNTIL=$(date -v+1d +"%Y-%m-%dT00:00Z")
fi

AZURE_STORAGE_ACCOUNT=modelstaging
AZURE_STORAGE_SAS_TOKEN=$(az storage container generate-sas \
    --subscription FortiSubscription \
    -n collected \
    --account-name "$AZURE_STORAGE_ACCOUNT" \
    --https-only \
    --permissions lr \
    --expiry "$VALID_UNTIL" \
    -otsv \
)

export AZURE_STORAGE_ACCOUNT=$AZURE_STORAGE_ACCOUNT
export AZURE_STORAGE_SAS_TOKEN=$AZURE_STORAGE_SAS_TOKEN
