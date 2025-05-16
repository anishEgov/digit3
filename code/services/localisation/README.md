# Localization Service (Go)

A Go-based implementation of the DIGIT localization service using the Gin framework. This service provides locale-specific components and translates text for applications, similar to the original Java Spring Boot-based egov-localization service.

## Features

- Store and retrieve locale-specific messages with key-value pairs
- Multi-tenant and multi-language support
- Efficient caching with Redis
- PostgreSQL database for persistent storage
- Clean architecture with separation of concerns

## API Endpoints

### Upsert Messages

- **Endpoint**: `/localization/messages/v1/_upsert`
- **Method**: POST
- **Description**: Creates or updates localization messages
- **Request Body**: JSON structure with tenantId and array of messages
- **Response**: JSON with the newly created/updated messages

### Search Messages

- **Endpoint**: `/localization/messages/v1/_search`
- **Method**: POST
- **Description**: Searches for localization messages based on various criteria
- **Request Body**: JSON structure with tenantId, module, locale, and optional codes
- **Response**: JSON with matching messages

## Configuration

The service can be configured using environment variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| SERVER_PORT | Port on which the server runs | 8088 |
| DB_HOST | PostgreSQL database host | localhost |
| DB_PORT | PostgreSQL database port | 5432 |
| DB_USER | PostgreSQL database username | postgres |
| DB_PASSWORD | PostgreSQL database password | postgres |
| DB_NAME | PostgreSQL database name | localization |
| DB_SSL_MODE | PostgreSQL SSL mode | disable |
| REDIS_HOST | Redis server host | localhost |
| REDIS_PORT | Redis server port | 6379 |
| REDIS_PASSWORD | Redis server password | (empty) |
| REDIS_DB | Redis database index | 0 |
| CACHE_EXPIRATION | Cache expiration duration | 24h |

## Development

### Prerequisites

- Go 1.16 or later
- PostgreSQL 12 or later
- Redis 6 or later

### Installation

1. Clone the repository
   ```
   git clone https://github.com/yourusername/localisationgo.git
   cd localisationgo
   ```

2. Install dependencies
   ```
   go mod download
   ```

3. Build the service
   ```
   go build -o localization-service ./cmd/server
   ```

4. Run the service
   ```
   ./localization-service
   ```

### Testing

The localization service includes comprehensive unit and integration tests to ensure functionality remains intact when changes are made.

### Running Tests

Run all tests with:

```bash
go test ./...
```

Run specific test packages:

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./tests/...
```

### Test Coverage

Generate a test coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

- **Unit Tests**: Each component has corresponding tests in the same package with the `_test.go` suffix
  - Service tests: `internal/core/services/messageservice_test.go`
  - Repository tests: `internal/repositories/postgres/messagerepository_test.go`
  - Cache tests: `internal/platform/cache/rediscache_test.go`
  - Handler tests: `internal/handlers/messagehandler_test.go`

- **Integration Tests**: End-to-end tests that test the complete flow
  - `tests/integration_test.go`: Tests the complete flow from HTTP request to database and back

### Test Dependencies

The tests use the following testing libraries:
- `github.com/stretchr/testify`: For assertions and mocks
- `github.com/DATA-DOG/go-sqlmock`: For SQL mocking
- `github.com/alicebob/miniredis/v2`: For Redis mocking
- `github.com/mattn/go-sqlite3`: For in-memory SQLite database in integration tests

## Example Usage

### Upsert Messages

```bash
curl --location 'http://localhost:8088/localization/messages/v1/_upsert' \
--header 'Content-Type: application/json' \
--data '{
    "RequestInfo": {
        "apiId": "emp",
        "ver": "1.0",
        "action": "create",
        "did": "1",
        "key": "abcdkey",
        "msgId": "20170310130900",
        "requesterId": "rajesh",
        "authToken": "0cfe07e1-94b5-4f50-a7a0-c7c186feb9d5",
        "userInfo": {
            "id": 128
        }
    },
    "tenantId": "DEFAULT",
    "messages": [
      {
            "code": "ComplaintsInbox",
            "message": "Complaints Inbox",
            "module": "digit-ui",
            "locale": "en_IN"
        }
    ]
}'
```

### Search Messages

```bash
curl --location 'http://localhost:8088/localization/messages/v1/_search' \
--header 'Content-Type: application/json' \
--data '{
    "RequestInfo": {
        "apiId": "emp",
        "ver": "1.0",
        "action": "search",
        "did": "1",
        "key": "abcdkey",
        "msgId": "20170310130900",
        "requesterId": "rajesh",
        "authToken": "0cfe07e1-94b5-4f50-a7a0-c7c186feb9d5",
        "userInfo": {
            "id": 128
        }
    },
    "tenantId": "DEFAULT",
    "module": "digit-ui",
    "locale": "en_IN"
}'
```

## License

MIT 