FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o bootstrap main.go

FROM public.ecr.aws/lambda/go:1

COPY --from=builder /app/bootstrap ${LAMBDA_TASK_ROOT}

CMD ["bootstrap"]
