FROM golang:1.26-alpine3.21 AS builder

# Install build dependencies
RUN apk add --no-cache git make protobuf-dev gcc musl-dev

WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate proto files
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Build broker binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -o /build/bin/broker ./cmd/broker

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/bin/broker /app/broker

# Create data directory
RUN mkdir -p /app/data

# Expose gRPC port for e2e network
EXPOSE 9091

# Set environment variables (can be overridden)
ENV BROKER_HOST=0.0.0.0 \
    BROKER_PORT=9091 \
    BROKER_ID=1 \
    COORDINATOR_URL=localhost:2379 \
    DATA_DIR=/app/data

# Run broker with gRPC port 9091
ENTRYPOINT ["/app/broker"]
CMD ["--host", "0.0.0.0", "--port", "9091", "--id", "1"]
