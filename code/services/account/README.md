# Account Service

A microservice responsible for managing user accounts and authentication in the Digit platform.

## Overview

This service provides account management functionality, including user registration, authentication, and profile management. It's built using Go and uses PostgreSQL as its database.

## Features

- User account management
- Authentication
- Profile management
- Database migrations support

## Tech Stack

- Go 1.22 
- PostgreSQL
- Chi Router
- Docker
- Google Cloud Run (Deployment)

## Prerequisites

- Go 1.22 or later
- Docker
- PostgreSQL
- Make

## Getting Started

### Local Development

1. Clone the repository
2. Set up your environment variables:
   ```bash
   export DB_URL=postgres://postgres:password@localhost:5432/digit?sslmode=disable
   ```

3. Run database migrations:
   ```bash
   make migrate-up
   ```

4. Build and run the service:
   ```bash
   go build -o account ./cmd/main.go
   ./account
   ```

### Docker

To run the service using Docker:

```bash
docker build -t account .
docker run -p 8080:8080 account
```

## Database Migrations

The service uses the `migrate` tool for database migrations:

- Create a new migration:
  ```bash
  make create-migration name=migration_name
  ```

- Apply migrations:
  ```bash
  make migrate-up
  ```

- Rollback migrations:
  ```bash
  make migrate-down
  ```

## API Documentation

The service exposes a REST API on port 8080. Detailed API documentation can be found in the API specification.

## Deployment

The service is automatically deployed to Google Cloud Run when changes are pushed to the main branch. The deployment process includes:

1. Building a Docker image
2. Pushing the image to Google Artifact Registry
3. Deploying to Cloud Run

## Environment Variables

- `DB_URL`: PostgreSQL connection string
- `PORT`: Service port (default: 8080)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[Add your license information here] # retry deploy
