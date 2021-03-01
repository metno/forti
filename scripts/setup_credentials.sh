#!/bin/bash

VALID_UNTIL=$(date -d "1 day" +"%Y-%m-%dT00:00Z")

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

export AZURE_STORAGE_ACCOUNT
export AZURE_STORAGE_SAS_TOKEN

