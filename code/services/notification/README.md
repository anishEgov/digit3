# Notification Service

The Notification Service is a robust and scalable microservice designed to handle all aspects of user notifications within a multi-tenant architecture. It provides a comprehensive solution for managing notification templates, sending emails and SMS messages, and enriching notification content with dynamic data.

## Features

- **Template Management**: CRUD operations for notification templates (both Email and SMS).
- **Multi-Channel Notifications**: Supports sending notifications via Email and SMS.
- **Synchronous and Asynchronous Sending**:
  - **Sync**: Send notifications on-demand via REST API endpoints.
  - **Async**: Consume messages from Kafka or Redis message brokers to send notifications asynchronously.
- **Dynamic Content Enrichment**: Integrates with a `template-config` service to enrich notification payloads with additional data before rendering.
- **Template Preview**: Preview rendered templates with or without data enrichment to ensure correctness.
- **Multi-Tenancy**: Supports multi-tenant environments using the `X-Tenant-ID` header.
- **Configurable Providers**: Easily configure SMTP providers for email and providers for SMS.
- **Database Migrations**: Automated database migration support.
- **File Store Integration**: Upload files and include them as attachments in notifications.

## How it Works

### 1. Template Management

The service allows you to create, read, update, and delete notification templates. Each template has a `type` (EMAIL or SMS), `subject` (for emails), `content`, and can be designated as HTML or plain text.

### 2. Notification Sending

You can send notifications in two ways:

- **Synchronous API Calls**: For immediate, on-demand notifications, you can use the `/email/send` and `/sms/send` endpoints.
- **Asynchronous Message Handling**: For high-throughput or non-critical notifications, the service can consume events from a message broker (Kafka or Redis). It listens on specific topics for email and SMS requests, processes them, and sends the notifications.

### 3. Payload Enrichment

When sending a notification or previewing a template, you can enable the `enrich` option. If enabled, the service will:
1. Call the `template-config` service with the initial payload.
2. The `template-config` service enriches the payload by adding more data (e.g., user profile information, account details).
3. The enriched payload is then used to render the notification template, resulting in highly personalized content.

## API Endpoints

All endpoints are prefixed with the `SERVER_CONTEXT_PATH` (default: `/notification`).

### Template Management

- `POST /template`: Create a new notification template.
- `PUT /template`: Update an existing notification template.
- `GET /template`: Search for notification templates.
- `DELETE /template`: Delete a notification template.
- `POST /template/preview`: Preview a rendered template.

### Notification Sending

- `POST /email/send`: Send an email synchronously.
- `POST /sms/send`: Send an SMS synchronously.

### Example: Create and Preview a Template

1.  **Create a template:**
    ```bash
    curl -X POST http://localhost:8080/notification/template \
      -H "Content-Type: application/json" \
      -H "X-Tenant-ID: tenant1" \
      -d '{
        "templateId": "welcome-email",
        "version": "1.0",
        "type": "EMAIL",
        "subject": "Welcome {{.userName}}!",
        "content": "<h1>Hello, {{.userName}}!</h1><p>Welcome to our platform.</p>",
        "isHTML": true
      }'
    ```

2.  **Preview the template:**
    ```bash
    curl -X POST http://localhost:8080/notification/template/preview \
      -H "Content-Type: application/json" \
      -H "X-Tenant-ID: tenant1" \
      -d '{
        "templateId": "welcome-email",
        "version": "1.0",
        "payload": {
          "userName": "JohnDoe"
        }
      }'
    ```

## Configuration

The service is configured using environment variables.

| Environment Variable          | Default Value                               | Description                                                 |
| ----------------------------- | ------------------------------------------- | ----------------------------------------------------------- |
| `HTTP_PORT`                   | `8080`                                      | Port for the HTTP server.                                   |
| `SERVER_CONTEXT_PATH`         | `/notification`                             | Base path for the API endpoints.                            |
| `DB_HOST`                     | `localhost`                                 | Database host.                                              |
| `DB_PORT`                     | `5432`                                      | Database port.                                              |
| `DB_USER`                     | `postgres`                                  | Database username.                                          |
| `DB_PASSWORD`                 | `postgres`                                  | Database password.                                          |
| `DB_NAME`                     | `notification_template`                     | Database name.                                              |
| `DB_SSL_MODE`                 | `disable`                                   | Database SSL mode.                                          |
| `MIGRATION_ENABLED`           | `false`                                     | Enable/disable database migrations on startup.              |
| `MIGRATION_SCRIPT_PATH`       | `./migrations`                              | Path to the database migration scripts.                     |
| `TEMPLATE_CONFIG_HOST`        | `http://localhost:8082`                     | Base URL of the template-config service.                    |
| `TEMPLATE_CONFIG_PATH`        | `/template-config/v1/render`                | Path for the template-config render endpoint.               |
| `FILESTORE_HOST`              | `http://localhost:8083`                     | Base URL of the filestore service.                          |
| `FILESTORE_PATH`              | `/filestore/v1/upload`                      | Path for the filestore upload endpoint.                     |
| `SMTP_HOST`                   | `smtp.gmail.com`                            | SMTP server host.                                           |
| `SMTP_PORT`                   | `587`                                       | SMTP server port.                                           |
| `SMTP_USERNAME`               | `username`                                  | SMTP username.                                              |
| `SMTP_PASSWORD`               | `password`                                  | SMTP password.                                              |
| `SMTP_FROM_ADDRESS`           | `notification@example.com`                  | Default "from" email address.                               |
| `SMTP_FROM_NAME`              | `Notification Service`                      | Default "from" name.                                        |
| `SMS_PROVIDER_URL`            | `https://smscountry.com/api/v3/sendsms/plain` | URL of the SMS provider.                                    |
| `SMS_PROVIDER_USERNAME`       | `username`                                  | SMS provider username.                                      |
| `SMS_PROVIDER_PASSWORD`       | `password`                                  | SMS provider password.                                      |
| `SMS_PROVIDER_CONTENT_TYPE`   | `application/x-www-form-urlencoded`         | Content type for the SMS provider API.                      |
| `MESSAGE_BROKER_ENABLED`      | `false`                                     | Enable/disable the message broker consumer.                 |
| `MESSAGE_BROKER_TYPE`         | `kafka`                                     | Type of message broker (`kafka` or `redis`).                |
| `KAFKA_BROKERS`               | `localhost:9092`                            | Comma-separated list of Kafka brokers.                      |
| `KAFKA_CONSUMER_GROUP`        | `notification-consumer-group`               | Kafka consumer group ID.                                    |
| `REDIS_ADDR`                  | `localhost:6379`                            | Redis server address.                                       |
| `REDIS_PASSWORD`              | `""`                                        | Redis password.                                             |
| `REDIS_DB`                    | `0`                                         | Redis database number.                                      |
| `EMAIL_TOPIC`                 | `notification-email`                        | Topic/channel for email notifications.                      |
| `SMS_TOPIC`                   | `notification-sms`                          | Topic/channel for SMS notifications.                        |

## Running the Service

1.  **Set up the environment variables** as described in the configuration section.
2.  **Ensure the database and other dependencies (like Kafka/Redis, template-config service) are running.**
3.  **Start the service:**
    ```bash
    go run cmd/server/main.go
    ```
