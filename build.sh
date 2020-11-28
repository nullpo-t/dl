#!/bin/sh

# load .env
set -o allexport; source .env; set +o allexport

docker build -t gcr.io/$DL_GCP_ID/dl:$DL_VERSION .
docker push gcr.io/$DL_GCP_ID/dl:$DL_VERSION

