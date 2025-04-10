#!/bin/bash

# ---- CONFIGURE ----
PROJECT_ID="digitnxt-456204"
REGION="us-central1"
SERVICE_ACCOUNT_NAME="github-deployer"
ARTIFACT_REGISTRY="digit-docker-repo"
# -------------------

echo "🔹 Using existing project: $PROJECT_ID"
gcloud config set project "$PROJECT_ID"

echo "🔹 Enabling required services..."
gcloud services enable \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  run.googleapis.com \
  iam.googleapis.com

echo "🔹 Creating service account..."
gcloud iam service-accounts create "$SERVICE_ACCOUNT_NAME" \
  --description="GitHub Actions deployer" \
  --display-name="GitHub Deployer"

SA_EMAIL="$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com"

echo "🔹 Assigning roles to service account..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$SA_EMAIL" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$SA_EMAIL" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$SA_EMAIL" \
  --role="roles/iam.serviceAccountUser"

echo "🔹 Creating service account key file..."
gcloud iam service-accounts keys create "./gcp-key-$PROJECT_ID.json" \
  --iam-account="$SA_EMAIL"

echo "🔹 Creating Docker Artifact Registry (if not exists)..."
gcloud artifacts repositories create "$ARTIFACT_REGISTRY" \
  --repository-format=docker \
  --location="$REGION" \
  --description="Docker registry for DIGIT microservices" || echo "✅ Skipping, already exists."

echo "✅ All set!"
echo "➡️  Encode the key for GitHub Actions with:"
echo "cat gcp-key-$PROJECT_ID.json | base64"