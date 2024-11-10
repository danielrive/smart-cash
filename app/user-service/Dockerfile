# Use a Golang base image for the build stage
FROM golang:1.22-alpine AS build

# Copy the Go modules files and download dependencies

WORKDIR /app

COPY go.mod .

COPY go.sum .

#RUN go mod download

# Copy the application source code to the container

COPY  utils/*.go ./utils/
COPY  user-service/models/* ./user-service/models/
COPY  user-service/internal ./user-service/internal
COPY  user-service/*.go ./user-service/

# Build the Go application statically

WORKDIR /app/user-service

# Build the Go application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/user_service .

# Use a minimal base image for the runtime stage
FROM gcr.io/distroless/base

# Copy the binary from the build stage
COPY --from=build  /bin/user_service /bin/user_service

# Expose the port on which the application will listen
EXPOSE 8181

# Set the command to run the application when the container starts.

ENTRYPOINT ["/bin/user_service"]
