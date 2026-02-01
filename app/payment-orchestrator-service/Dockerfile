# Use a Golang base image for the build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

COPY utils ./utils
COPY payment-orchestrator-service/models/* ./payment-orchestrator-service/models/
COPY payment-orchestrator-service/internal ./payment-orchestrator-service/internal
COPY payment-orchestrator-service/*.go ./payment-orchestrator-service/

# Build the Go application statically
WORKDIR /app/payment-orchestrator-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/payment_orchestrator_service .

# Use a minimal base image for the runtime stage
FROM gcr.io/distroless/base

# Copy the binary from the build stage
COPY --from=build /bin/payment_orchestrator_service /bin/payment_orchestrator_service

# Expose port for health checks (optional, but good practice)
EXPOSE 8080

# Set the command to run the application when the container starts
ENTRYPOINT ["/bin/payment_orchestrator_service"]
