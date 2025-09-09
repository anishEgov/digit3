# syntax=docker/dockerfile:1.4

# --- Build stage ---
FROM golang:1.22.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build for ARM64 architecture (matches your cluster nodes)
RUN CGO_ENABLED=0 GOOS=linux  go build -o tenant-management-go cmd/server/main.go

# --- Final image ---
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/tenant-management-go ./tenant-management-go
# Copy the realm config file to ensure it's available
COPY internal/keycloak/realm_config.json ./realm_config.json
COPY internal/db ./db
EXPOSE 8080
ENTRYPOINT ["./tenant-management-go"] 