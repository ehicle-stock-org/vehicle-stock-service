# Build stage
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o vehicle-stock-service main.go

# Final image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/vehicle-stock-service .
EXPOSE 8080
CMD ["/app/vehicle-stock-service"]
