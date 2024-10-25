# Use a Golang base image for the build stage
FROM golang:1.22-alpine AS build

# Copy the Go modules files and download dependencies

WORKDIR /app

COPY go.mod .

COPY go.sum .

#RUN go mod download

# Copy the application source code to the container

COPY  utils/*.go ./utils/
COPY  bank-service/internal/common/* ./bank-service/internal/common/
COPY  bank-service/internal/handler/* ./bank-service/internal/handler/
COPY  bank-service/models/* ./bank-service/models/
COPY  bank-service/internal/repositories/* ./bank-service/internal/repositories/
COPY  bank-service/internal/service/* ./bank-service/internal/service/
COPY  bank-service/main.go ./bank-service/main.go

# Build the Go application statically

WORKDIR /app/bank-service
# Build the Go application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/bank_service main.go

# Use a minimal base image for the runtime stage
FROM gcr.io/distroless/base

# Copy the binary from the build stage
COPY --from=build  /bin/bank_service /bin/bank_service

# Expose the port on which the application will listen
EXPOSE 8282

# Set the command to run the application when the container starts
ENTRYPOINT ["/bin/bank_service"]