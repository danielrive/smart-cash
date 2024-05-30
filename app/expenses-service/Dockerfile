# Use a Golang base image for the build stage
FROM golang:1.22-alpine AS build

# Copy the Go modules files and download dependencies

WORKDIR /app

COPY go.mod .

COPY go.sum .

#RUN go mod download

# Copy the application source code to the container

COPY  utils/*.go ./utils/
COPY  expenses-service/internal/common/* ./expenses-service/internal/common/
COPY  expenses-service/internal/handler/* ./expenses-service/internal/handler/
COPY  expenses-service/internal/models/* ./expenses-service/internal/models/
COPY  expenses-service/internal/repositories/* ./expenses-service/internal/repositories/
COPY  expenses-service/internal/service/* ./expenses-service/internal/service/
COPY  expenses-service/main.go ./expenses-service/main.go

# Build the Go application statically

WORKDIR /app/expenses-service
# Build the Go application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/expenses_service main.go

# Use a minimal base image for the runtime stage
FROM gcr.io/distroless/base

# Copy the binary from the build stage
COPY --from=build  /bin/expenses_service /bin/expenses_service

# Expose the port on which the application will listen
EXPOSE 8282

# Set the command to run the application when the container starts
ENTRYPOINT ["/bin/expenses_service"]
