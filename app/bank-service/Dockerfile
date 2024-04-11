# Use a Golang base image for the build stage
FROM golang:1.22-alpine AS build

# Copy the Go modules files and download dependencies

WORKDIR app

COPY ../../go.mod ./

RUN go mod download

# Copy the application source code to the container
COPY . ./

# Build the Go application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/bank_service main.go

# Use a minimal base image for the runtime stage
FROM gcr.io/distroless/base

# Copy the binary from the build stage
COPY --from=build  /bin/bank_service /bin/bank_service

# Expose the port on which the application will listen
EXPOSE 8282

# Set the command to run the application when the container starts
CMD ["/bin/bank_service"]
