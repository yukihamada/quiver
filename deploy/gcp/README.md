# QUIVer GCP Deployment

This directory contains Terraform configuration to deploy QUIVer nodes on Google Cloud Platform.

## Prerequisites

1. Install Terraform:
   ```bash
   brew install terraform
   ```

2. Install Google Cloud SDK:
   ```bash
   brew install google-cloud-sdk
   ```

3. Authenticate with GCP:
   ```bash
   gcloud auth application-default login
   ```

4. Create a GCP project and enable required APIs:
   ```bash
   gcloud projects create quiver-network-prod
   gcloud config set project quiver-network-prod
   gcloud services enable compute.googleapis.com
   ```

## Deployment

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Plan the deployment:
   ```bash
   terraform plan -var="project_id=quiver-network-prod"
   ```

3. Deploy the infrastructure:
   ```bash
   terraform apply -var="project_id=quiver-network-prod"
   ```

## Architecture

The deployment creates:
- 1 Bootstrap node (e2-medium)
- 2 Provider nodes (n1-standard-2) with Ollama and llama3.2
- 1 Gateway node (e2-medium)
- 1 Realtime stats node (e2-micro) with WebSocket support

## Access

After deployment, you can access:
- WebSocket stats: `ws://<stats_ip>:8087/ws`
- HTTP stats API: `http://<stats_ip>/api/stats`
- Gateway API: `http://<gateway_ip>:8081`

## Real-time Monitoring

The stats node provides real-time network statistics via WebSocket. Connect to see:
- Live node count
- Node locations (GCP, AWS, local)
- Network health
- Individual node details

## Cleanup

To destroy all resources:
```bash
terraform destroy -var="project_id=quiver-network-prod"
```