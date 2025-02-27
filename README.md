<!-- prettier-ignore -->
<div align="center">

<img src="./docs/images/icon.png" alt="" align="center" height="64" /> <img src="./docs/images/go-logo.png" alt="" align="center" height="104" />


# DeepSeek-R1 Go Starter

![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)
![Go version](https://img.shields.io/badge/Go->=1.21-00ADD8?style=flat-square&logo=go&logoColor=white)

This sample demonstrates how to build a web chat application using Go that integrates with Azure OpenAI's DeepSeek-R1 model. The application provides a web interface for interacting with the model, deployed securely using Azure Container Registry (ACR) and Azure App Service.

:star: If you like this sample, star it on GitHub — it helps a lot!

</div>

## Overview

This application showcases how to build an AI chat application using Go, with features including:
- Web-based chat interface
- Docker containerization
- Azure infrastructure deployment using Terraform
- Automatic deployment pipeline
- Rate limiting and concurrent request handling

## Features

- **Web Interface**: Modern web-based chat UI for interacting with DeepSeek-R1 deployed in Azure AI Foundry
- **Docker Support**: Containerized application for consistent deployment
- **Azure Integration**: 
  - Azure Container Registry for image management
  - Azure App Service for hosting
  - Azure AI Foundry integration with DeepSeek-R1 reasoning model
- **Security Features**:
  - Environment variable configuration
  - Rate limiting
  - Request concurrency management
- **Infrastructure as Code**: Complete Terraform configuration for Azure resources

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker
- Azure CLI
- Terraform CLI
- Azure subscription with permissions to create resources

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/Azure-Samples/deepseek-go
   cd deepseek-go
   ```

2. Create a .env file with required configuration:
   ```
   AZURE_OPENAI_API_KEY=<your-api-key>
   MODEL_DEPLOYMENT_URL=<your-model-url>
   MODEL_DEPLOYMENT_NAME=DeepSeek-R1
   AZURE_OPENAI_ENDPOINT=<your-model-endpoint>
   ACR_NAME=<your-acr-name>
   IMAGE_NAME=<your-image-name>
   TAG=latest
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

The application will start on http://localhost:3000

### Docker Development

Build and run the application using Docker:

```bash
docker build -t deepseek-go .
docker run -p 3000:3000 --env-file .env deepseek-go
```

### GitHub Codespaces Development

1. Click the "Code" button on the repository and select "Open with Codespaces"
2. Wait for the devcontainer to build and initialize
3. Copy `.env.example` to `.env` and update with your Azure credentials:
   ```bash
   cp .env.example .env
   ```
4. Run the application:
   ```bash
   go run main.go
   ```

The application will be available on port 3000, which is automatically forwarded by Codespaces.

## Deployment

### Azure Infrastructure Setup

1. Initialize Terraform:
   ```bash
   cd terraform
   terraform init
   ```

2. Deploy the infrastructure:
   ```bash
   terraform apply
   ```

The deployment will:
- Create a resource group
- Set up Azure Container Registry
- Configure Azure App Service
- Deploy the application container

### Automatic Deployment

The application includes a deployment script (deploy.sh) that:
- Logs into Azure Container Registry
- Builds the Docker image
- Tags the image
- Pushes to ACR
- Triggers a web app update

## Project Structure

```
├── main.go              # Application entry point
├── Dockerfile           # Container configuration
├── deploy.sh            # Deployment script
├── static/              # Web interface assets
│   ├── index.html
│   └── styles.css
└── terraform/           # Infrastructure as code
    ├── main.tf
    ├── variables.tf
    ├── outputs.tf
    └── providers.tf
```

## Configuration

The application uses environment variables for configuration:

- `AZURE_OPENAI_API_KEY`: Your Azure OpenAI API key
- `MODEL_DEPLOYMENT_URL`: The deployment URL for your model
- `MODEL_DEPLOYMENT_NAME`: The model name (defaults to "DeepSeek-R1")
- `AZURE_OPENAI_ENDPOINT`: The endpoint for Azure OpenAI
- `ACR_NAME`: The name of your Azure Container Registry
- `IMAGE_NAME`: The name of the Docker image
- `TAG`: The image tag (defaults to "latest")

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
