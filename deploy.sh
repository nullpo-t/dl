#!/bin/sh

set -x

# load .env
set -o allexport; source ./.env; set +o allexport

# set GCP project
gcloud config set project $DL_GCP_ID

# build and push container image
docker build -t gcr.io/$DL_GCP_ID/$DL_APP_NAME:$DL_VERSION .
docker push gcr.io/$DL_GCP_ID/$DL_APP_NAME:$DL_VERSION

# deploy
# sets --max-instances=1 to avoid data conflict; **this app is not safe to run multiple instances**
gcloud run deploy $DL_APP_NAME \
  --image gcr.io/$DL_GCP_ID/$DL_APP_NAME:$DL_VERSION \
  --platform managed \
  --memory=128Mi --cpu=1000m \
  --max-instances=1 \
  --set-env-vars=DL_GCP_ID=$DL_GCP_ID,DL_GCS_APP_BUCKET=$DL_GCS_APP_BUCKET,DL_GCS_DATA_BUCKET=$DL_GCS_DATA_BUCKET \
  --region=asia-northeast1 \
  --service-account=$DL_GCP_RUN_SERVICE_ACCOUNT \
  --allow-unauthenticated
