# Vehicle Stock Service

## Project Overview
This service manages vehicle stock data, parses vehicle subscription JSON, checks for active paid subscriptions, and safely handles optional fields. It integrates Kafka, MongoDB, Stripe, and is cloud-native deployable. Designed for enterprise use (Toyota), it supports cloud deployment via Helm/Minikube and includes full audit documentation.

## Implementation & Used Technologies
- **Language:** Go (Golang)
- **REST API:** Gorilla Mux
- **Messaging:** Kafka (Confluent Cloud supported)
- **Database:** MongoDB (Atlas supported)
- **Payment Processor:** Stripe Go SDK (manual capture/hold)
- **Cloud Deployment:** Docker, Helm, Minikube
- **Testing:** Go test, Testify
- **Code Quality:** SonarQube

## How to Build & Run Locally
1. Download dependencies:
   ```
   go mod tidy
   ```
2. Build all packages:
   ```
   go build ./...
   ```
3. Run the main service:
   ```
   go run main.go
   ```

## API Endpoints
- **GET /getstock**
  - Mandatory header parameters: `startDate`, `endDate`
  - Returns bid/ask prices for the given dates and their difference.
- **POST /holdpayment**
  - Places a hold on a payment method using Stripe (manual capture).

## Deployment (Minikube & Helm)
1. Start Minikube:
   ```
   minikube start
   ```
2. Configure shell for Minikube Docker:
   - PowerShell:
     ```
     minikube -p minikube docker-env --shell powershell | Invoke-Expression
     ```
   - Command Prompt:
     ```
     @FOR /f "tokens=*" %i IN ('minikube -p minikube docker-env --shell cmd') DO @%i
     ```
3. Build Docker image inside Minikube:
   ```
   docker build -t vehicle-stock-service:latest .
   ```
4. Install Helm chart:
   ```
   helm install vehicle-stock-service ./charts/vehicle-stock-service
   ```
5. Check deployment status:
   ```
   kubectl get pods
   kubectl get services
   ```

## Configuration
- See `config.yaml` for service configuration.
- MongoDB/Kafka/Stripe credentials should be set as environment variables or in config files.

## SonarQube Quality Gate
- Ensure coverage meets your organization's requirements.

## Full Dependency List
- See `RESULTS.md` for a complete list of downloaded Go dependencies for audit and compliance.
