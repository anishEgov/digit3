# Use an official Go build image
FROM golang:1.22 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app with correct binary name
RUN go build -o account ./cmd/main.go

# --- Final stage ---
FROM alpine:latest

WORKDIR /root

# Copy built binary from builder
COPY --from=builder /app/account .

# Expose service port
EXPOSE 8080

# Run the service
ENTRYPOINT ["./account"]