
# Vehicle Stock Service

## Overview
Vehicle Stock Service is a cloud-native, enterprise-grade Go application for managing vehicle stock data. It parses vehicle subscription JSON, checks for active paid subscriptions, and integrates with Kafka, MongoDB, and Stripe. Designed for scalable deployments, it supports Docker, Helm, and Minikube, and follows modern coding standards.

## Features
- REST API with Gorilla Mux
- Kafka publisher/consumer for stock data (Confluent Cloud compatible)
- MongoDB persistence (Atlas Community supported)
- Stripe payment hold (manual capture)
- Configurable via `config.json`, environment variables, or AWS Secrets Manager
- Cloud-native deployment: Docker, Helm, Minikube
- Comprehensive test coverage and SonarQube integration

## Configuration

### Local Development
- All service configuration is in `config.json`.
- Example:
   ```json
   {
      "kafka_brokers": ["localhost:9092"],
      "kafka_topic": "vehicle-stock",
      "mongo_uri": "mongodb://localhost:27017",
      "mongo_db": "vehicle_stock_db",
      "mongo_collection": "stock_data",
      "stripe_key": "sk_test_123"
   }
   ```

### Cloud/Production
- Set `ENV` to any value except `local`.
- Configuration is loaded from AWS Secrets Manager (recommended) or environment variables:
   - `KAFKA_BROKERS`, `KAFKA_TOPIC`, `MONGO_URI`, `MONGO_DB`, `MONGO_COLLECTION`, `STRIPE_KEY`
- AWS region is set via `AWS_REGION`.

## Build & Run

### Prerequisites
- Go 1.23+
- Docker (for containerization)
- Helm & Minikube (for Kubernetes)

### Local
```sh
go mod tidy
go build ./...
go run main.go
```

### Docker
```sh
docker build -t vehicle-stock-service:latest .
docker run -p 8080:8080 vehicle-stock-service:latest
```

### Kubernetes (Minikube)
```sh
minikube start
minikube -p minikube docker-env --shell powershell | Invoke-Expression
docker build -t vehicle-stock-service:latest .
helm install vehicle-stock-service ./charts/vehicle-stock-service
kubectl get pods
kubectl get services
```

## API Reference

### GET `/getstock`
- **Headers:** `startDate`, `endDate` (required)
- **Response:**
   - Bid/ask prices for each vehicle on the given dates
   - Price difference
   - Full vehicle payload

### POST `/holdpayment`
- **Body:**
   ```json
   {
      "amount": 1000,
      "currency": "usd",
      "payment_method": "pm_xxx"
   }
   ```
- **Response:** Stripe payment intent details

## Cloud Integration

- **Kafka:** Compatible with Confluent Cloud (set brokers in config)
- **MongoDB:** Compatible with Atlas Community (set URI in config)
- **Stripe:** Uses Stripe Go SDK, manual capture for payment holds
- **AWS Secrets Manager:** Secure config loading for cloud deployments

## Testing & Quality
- Run all tests:
   ```sh
   go test ./...
   ```
- SonarQube integration for code quality (see `RESULTS.md`)

## Security & Compliance
- No hardcoded credentials
- All secrets loaded via config file, env vars, or AWS Secrets Manager
- See `RESULTS.md` for dependency audit

## Contributing
- Follow Go best practices and cloud-native standards
- All code changes require tests and documentation

## License
See LICENSE file for details.
