# -------------------------
# Build stage
# -------------------------
FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o idgen ./cmd/main.go

# -------------------------
# Runtime stage
# -------------------------
FROM alpine:3.19

WORKDIR /opt/digit

COPY --from=build /app/idgen .
COPY --from=build /app/internal/migrations ./internal/migrations

RUN chmod +x /opt/digit/idgen

EXPOSE 8080
ENTRYPOINT ["./idgen"]
