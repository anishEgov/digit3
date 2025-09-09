# syntax=docker/dockerfile:1.4

# --- Build stage ---
    FROM golang:1.23-alpine AS builder
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .
    # Build for ARM64 architecture (matches your cluster nodes)
    RUN CGO_ENABLED=0 GOOS=linux go build -o boundary-go main.go
    
    # --- Final image ---
    FROM alpine:3.19
    WORKDIR /app
    COPY --from=builder /app/boundary-go ./boundary-go
    COPY db ./db
    EXPOSE 8080
    ENTRYPOINT ["./boundary-go"]