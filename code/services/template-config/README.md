# Template Config Service

The Template Config Service is a robust and scalable microservice that manages **template configurations** and enriches data dynamically using external API calls and JSONPath field mapping. It provides the foundation for dynamic notification rendering by transforming and enriching payloads before templates are applied.

## Features

- **Template Config Management**: CRUD operations for template configurations.
- **Data Transformation**: Extract and map fields from JSON payloads using JSONPath.
- **API Enrichment**: Make parallel external API calls and map responses into the payload.
- **Validation**: Ensures template configurations are well-formed before persisting.
- **Error Handling**: Detailed error reporting for invalid configs and failed API calls.
- **Multi-Tenancy**: Tenant-based isolation using the `X-Tenant-ID` header.
- **Database Migrations**: Automated database migration support.

## How it Works

### 1. Template Config Management
You can create, update, search, and delete template configurations. Each config defines:
- Field mappings from payload → enriched data.
- External API calls (method, endpoint, path, params).
- Response mappings from API → enriched data.

### 2. Rendering with Enrichment
When you call the `/render` endpoint:
1. Payload data is transformed using field mappings.
2. External APIs (if configured) are called in parallel.
3. API responses are mapped into the payload.
4. The final enriched payload is returned, ready for use in notifications.

### 3. Error Handling
- If one API call fails, the others still continue.
- Errors are collected and returned in the response.
- The service responds with **422 Unprocessable Entity** when enrichment fails partially.

## API Endpoints

All endpoints are prefixed with the `SERVER_CONTEXT_PATH` (default: `/template-config/v1`).

### Template Config Management

- `POST /config` → Create a new template configuration.  
- `PUT /config` → Update an existing template configuration.  
- `GET /config` → Search template configurations (by `templateId`, `version`, or `uuids`).  
- `DELETE /config` → Delete a template configuration.  

### Rendering

- `POST /render` → Render and enrich a payload using the specified template configuration.

### Example: Render with Enrichment

1. **Create a template config:**
   ```bash
   curl -X POST http://localhost:8082/template-config/v1/config \
     -H "Content-Type: application/json" \
     -H "X-Tenant-ID: tenant1" \
     -d '{
       "templateId": "user-profile",
       "version": "1.0",
       "fieldMapping": {
         "userName": "$.payload.user.name",
         "userEmail": "$.payload.user.email"
       },
       "apiMapping": [
         {
           "method": "GET",
           "endpoint": {
             "base": "https://api.example.com",
             "path": "/users/{{userId}}",
             "pathParams": {
               "userId": "$.payload.user.id"
             }
           },
           "responseMapping": {
             "userStatus": "$.response.status"
           }
         }
       ]
     }'
   ```

1. **Render a payload:**
   ```bash
   curl -X POST http://localhost:8082/template-config/v1/render \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant1" \
  -d '{
    "templateId": "user-profile",
    "version": "1.0",
    "payload": {
      "user": {
        "id": "123",
        "name": "John Doe",
        "email": "john.doe@example.com"
      }
    }
  }'
 ```  

## Configuration

The service is configured using environment variables.

| Environment Variable    | Default Value         | Description                                 |
| ----------------------- | --------------------- | ------------------------------------------- |
| `HTTP_PORT`             | `8080`                | Port for the HTTP server.                   |
| `SERVER_CONTEXT_PATH`   | `/template-config/v1` | Base path for the API endpoints.            |
| `DB_HOST`               | `localhost`           | Database host.                              |
| `DB_PORT`               | `5432`                | Database port.                              |
| `DB_USER`               | `postgres`            | Database username.                          |
| `DB_PASSWORD`           | `postgres`            | Database password.                          |
| `DB_NAME`               | `template_config`     | Database name.                              |
| `DB_SSL_MODE`           | `disable`             | Database SSL mode.                          |
| `MIGRATION_SCRIPT_PATH` | `./migrations`        | Path to the database migration scripts.     |
| `MIGRATION_ENABLED`     | `false`                | Enable/disable database migrations startup. |

## Running the Service

1.  **Set up the environment variables** as described in the configuration section.
2.  **Ensure PostgreSQL is running and the database exists.**
3.  **Start the service:**
    ```bash
    go run cmd/server/main.go
    ```