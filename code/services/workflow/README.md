# Workflow Service (Go)

A Go-based implementation of the DIGIT workflow service using the Gin framework. This service provides stateful workflow management, process definitions, and automated state transitions for applications.

## Overview

**Service Name:** workflow-go

**Purpose:** Provides multi-tenant, stateful workflow management services for DIGIT applications with process definitions, state transitions, and automated escalation capabilities.

**Owner/Team:** DIGIT Platform Team

## Architecture

**Tech Stack:**
- Go 1.24
- Gin Web Framework
- PostgreSQL (via GORM and pgx)
- Protocol Buffers (gRPC ready)
- Docker

**Core Responsibilities:**
- Define and manage workflow processes with states and actions
- Track process instances as they move through workflow states
- Enforce business rules and guard conditions at state transitions
- Support parallel workflow execution with branch coordination
- Provide auto-escalation capabilities based on SLA breaches
- Multi-tenant support for different organizations
- REST API interface for workflow operations

**Dependencies:**
- PostgreSQL 15
- UUID extension for PostgreSQL

### Diagrams

#### High-level Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        C1[Mobile Apps]
        C2[Web Apps]
        C3[Other Services]
    end
    
    subgraph "API Gateway"
        GW[API Gateway]
    end
    
    subgraph "Workflow Service"
        subgraph "REST API"
            H1[Process Handler]
            H2[State Handler]
            H3[Action Handler]
            H4[Transition Handler]
            H5[Escalation Handler]
        end
        
        subgraph "Business Logic"
            S1[Process Service]
            S2[State Service]
            S3[Action Service]
            S4[Transition Service]
            S5[Escalation Service]
        end
        
        subgraph "Data Layer"
            R1[Process Repository]
            R2[State Repository]
            R3[Action Repository]
            R4[Instance Repository]
            R5[Escalation Repository]
        end
        
        subgraph "Infrastructure"
            M1[Migration Runner]
            CFG[Configuration]
            SG[Security Guards]
        end
    end
    
    subgraph "External Systems"
        DB[(PostgreSQL)]
    end
    
    C1 --> GW
    C2 --> GW
    C3 --> GW
    
    GW --> H1
    GW --> H2
    GW --> H3
    GW --> H4
    GW --> H5
    
    H1 --> S1
    H2 --> S2
    H3 --> S3
    H4 --> S4
    H5 --> S5
    
    S1 --> R1
    S2 --> R2
    S3 --> R3
    S4 --> R4
    S5 --> R5
    
    R1 --> DB
    R2 --> DB
    R3 --> DB
    R4 --> DB
    R5 --> DB
    
    M1 --> DB
```

## Features

- ✅ Define and manage workflow processes with states and actions
- ✅ Multi-tenant support with tenant isolation
- ✅ Stateful process instance tracking
- ✅ Guard condition enforcement (attribute validation, assignee checks)
- ✅ Parallel workflow execution with branch coordination
- ✅ Auto-escalation based on SLA breaches
- ✅ Clean architecture with separation of concerns
- ✅ REST API with JSON responses
- ✅ Database migrations with rollback support
- ✅ Comprehensive audit trail
- ✅ Docker containerization
- ✅ Multi-step process orchestration

## Installation & Setup

### Local Development (Manual Setup)

**Prerequisites:**
- Go 1.24+
- PostgreSQL 15 with UUID extension

**Steps:**

1. Clone and setup
   ```bash
   git clone https://github.com/yourusername/workflow-go.git
   cd workflow-go
   go mod download
   ```

2. Setup PostgreSQL database
   ```bash
   createdb workflow
   psql workflow -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
   ```

3. Run migrations
   ```bash
   go run ./cmd/server --migrate
   ```

4. Start service
   ```bash
   go run ./cmd/server
   ```

### Docker Production Setup

**Build the image:**
```bash
docker build -t workflow-go:latest .
```

**Run with environment variables:**
```bash
docker run -p 8081:8081 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  workflow-go:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default Value | Required |
|----------|-------------|---------------|----------|
| `SERVER_PORT` | Port for REST API server | `8081` | No |
| `DB_HOST` | PostgreSQL database host | `localhost` | Yes |
| `DB_PORT` | PostgreSQL database port | `5432` | No |
| `DB_USER` | PostgreSQL database username | `postgres` | No |
| `DB_PASSWORD` | PostgreSQL database password | `postgres` | Yes |
| `DB_NAME` | PostgreSQL database name | `postgres` | No |
| `RUN_MIGRATIONS` | Whether to run migrations on startup | `true` | No |
| `MIGRATION_PATH` | Path to migration files | `db/migration` | No |
| `MIGRATION_TIMEOUT` | Migration timeout duration | `5m` | No |

### Example .env file

```bash
# Server Configuration
SERVER_PORT=8081

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secure_password
DB_NAME=workflow

# Migration Configuration
RUN_MIGRATIONS=true
MIGRATION_PATH=db/migration
MIGRATION_TIMEOUT=5m
```

## API Reference

### REST API Endpoints

#### 1. Create Process
- **Endpoint**: `POST /workflow/v3/process`
- **Description**: Creates a new workflow process definition
- **Headers**: `X-Tenant-ID: {tenantId}`
- **Request Body**:
```json
{
  "name": "Application Review Process",
  "code": "APP_REVIEW",
  "description": "Process for reviewing applications",
  "version": "1.0",
  "sla": 1440
}
```
- **Response**: `201 Created` with created process

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Database

    Client->>Handler: POST /process
    Handler->>Service: CreateProcess(process)
    
    Service->>Service: Validate process data
    Service->>Repository: Create process
    Repository->>Database: INSERT process
    Database-->>Repository: Created process
    Repository-->>Service: Process data
    
    Service-->>Handler: Created process
    Handler-->>Client: 201 Created with process
```

#### 2. Get Process Definitions
- **Endpoint**: `GET /workflow/v3/process/definition`
- **Description**: Retrieves process definitions with states and actions
- **Query Parameters**:
  - `id` (optional, array)
  - `name` (optional, array)
- **Response**: `200 OK` with process definitions

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Database

    Client->>Handler: GET /process/definition
    Handler->>Service: GetProcessDefinitions(tenantId, ids, names)
    
    Service->>Repository: Query processes
    Repository->>Database: SELECT processes
    Database-->>Repository: Process data
    Repository-->>Service: Process list
    
    Service->>Repository: Query states for each process
    Repository->>Database: SELECT states
    Database-->>Repository: State data
    Repository-->>Service: State list
    
    Service->>Repository: Query actions for each state
    Repository->>Database: SELECT actions
    Database-->>Repository: Action data
    Repository-->>Service: Action list
    
    Service->>Service: Build process definitions
    Service-->>Handler: Process definitions
    Handler-->>Client: 200 OK with definitions
```

#### 3. Create State
- **Endpoint**: `POST /workflow/v3/process/{processId}/state`
- **Description**: Creates a new state within a process
- **Headers**: `X-Tenant-ID: {tenantId}`
- **Request Body**:
```json
{
  "code": "SUBMITTED",
  "name": "Application Submitted",
  "description": "Application has been submitted for review",
  "processId": "process-uuid",
  "sla": 60,
  "isInitial": true
}
```
- **Response**: `201 Created` with created state

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Database

    Client->>Handler: POST /process/{id}/state
    Handler->>Service: CreateState(state)
    
    Service->>Service: Validate state data
    Service->>Repository: Create state
    Repository->>Database: INSERT state
    Database-->>Repository: Created state
    Repository-->>Service: State data
    
    Service-->>Handler: Created state
    Handler-->>Client: 201 Created with state
```

#### 4. Create Action
- **Endpoint**: `POST /workflow/v3/state/{stateId}/action`
- **Description**: Creates a new action/transition between states
- **Headers**: `X-Tenant-ID: {tenantId}`
- **Request Body**:
```json
{
  "name": "APPROVE",
  "label": "Approve Application",
  "currentState": "state-uuid-1",
  "nextState": "state-uuid-2",
  "attributeValidation": {
    "attributes": {
      "role": ["REVIEWER", "ADMIN"],
      "department": ["IT", "HR"]
    },
    "assigneeCheck": true
  }
}
```
- **Response**: `201 Created` with created action

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Database

    Client->>Handler: POST /state/{id}/action
    Handler->>Service: CreateAction(action)
    
    Service->>Service: Validate action data
    Service->>Repository: Create attribute validation
    Repository->>Database: INSERT attribute_validation
    Database-->>Repository: Created validation
    Repository-->>Service: Validation data
    
    Service->>Repository: Create action
    Repository->>Database: INSERT action
    Database-->>Repository: Created action
    Repository-->>Service: Action data
    
    Service-->>Handler: Created action
    Handler-->>Client: 201 Created with action
```

#### 5. Process Transition
- **Endpoint**: `POST /workflow/v3/transition`
- **Description**: Transitions a process instance to the next state
- **Headers**: `X-Tenant-ID: {tenantId}`
- **Request Body**:
```json
{
  "processId": "process-uuid",
  "entityId": "application-123",
  "action": "APPROVE",
  "comment": "Application approved by reviewer",
  "assigner": "user-123",
  "assignees": ["user-456"],
  "attributes": {
    "role": ["REVIEWER"],
    "department": ["IT"]
  }
}
```
- **Response**: `202 Accepted` with updated process instance

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Guard
    participant Repository
    participant Database

    Client->>Handler: POST /transition
    Handler->>Service: Transition(processInstance)
    
    Service->>Repository: Get process instance
    Repository->>Database: SELECT process_instance
    Database-->>Repository: Instance data
    Repository-->>Service: Process instance
    
    Service->>Repository: Get action details
    Repository->>Database: SELECT action
    Database-->>Repository: Action data
    Repository-->>Service: Action details
    
    Service->>Guard: Validate guard conditions
    Guard->>Guard: Check attributes
    Guard->>Guard: Check assignee
    Guard-->>Service: Validation result
    
    alt Validation passed
        Service->>Repository: Update process instance
        Repository->>Database: UPDATE process_instance
        Database-->>Repository: Updated instance
        Repository-->>Service: Instance data
        
        Service-->>Handler: Updated instance
        Handler-->>Client: 202 Accepted with instance
    else Validation failed
        Service-->>Handler: Validation error
        Handler-->>Client: 400 Bad Request
    end
```

#### 6. Auto-Escalation
- **Endpoint**: `POST /workflow/v3/auto/{processCode}/_escalate`
- **Description**: Escalates process instances based on SLA breaches
- **Headers**: `X-Tenant-ID: {tenantId}`
- **Request Body**:
```json
{
  "attributes": {
    "role": ["ADMIN"],
    "department": ["IT"]
  }
}
```
- **Response**: `200 OK` with escalation results

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Database

    Client->>Handler: POST /auto/{code}/_escalate
    Handler->>Service: EscalateApplications(processCode, attributes)
    
    Service->>Repository: Get escalation configs
    Repository->>Database: SELECT escalation_configs
    Database-->>Repository: Config data
    Repository-->>Service: Escalation configs
    
    Service->>Repository: Find instances to escalate
    Repository->>Database: SELECT process_instances
    Database-->>Repository: Instance data
    Repository-->>Service: Instances to escalate
    
    loop For each instance
        Service->>Service: Apply escalation action
        Service->>Repository: Update instance
        Repository->>Database: UPDATE process_instance
        Database-->>Repository: Updated instance
    end
    
    Service-->>Handler: Escalation results
    Handler-->>Client: 200 OK with results
```

### Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-----------|
| 400 | BAD_REQUEST | Invalid request parameters |
| 401 | UNAUTHORIZED | Authentication required |
| 403 | FORBIDDEN | Insufficient permissions |
| 404 | NOT_FOUND | Resource not found |
| 409 | CONFLICT | Resource already exists |
| 422 | UNPROCESSABLE_ENTITY | Validation failed |
| 500 | INTERNAL_SERVER_ERROR | Server error |

## Observability

### Logging

**Format:** JSON structured logging with request correlation IDs

**Framework:** Standard Go log with context support

**Log Levels:** DEBUG, INFO, WARN, ERROR

**Example Log:**
```json
{
  "level": "INFO",
  "timestamp": "2024-01-15T10:30:45Z",
  "request_id": "req-123456",
  "tenant_id": "DEFAULT",
  "method": "POST",
  "path": "/workflow/v3/transition",
  "duration_ms": 120,
  "status_code": 202
}
```

### Metrics

**Framework:** Prometheus metrics exposed on `/metrics` endpoint

**Key Metrics:**
- `http_requests_total{path, method, status}` - Total HTTP requests
- `http_request_duration_seconds{path, method}` - Request duration histogram
- `db_connections_active` - Active database connections
- `process_instances_created_total` - Total process instances created
- `transitions_completed_total` - Total state transitions completed
- `escalations_triggered_total` - Total escalations triggered

### Tracing

**Framework:** OpenTelemetry with Jaeger integration

**Configuration:**
```bash
export OTEL_TRACES_EXPORTER=jaeger
export OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

**Trace Context:** Automatic trace propagation with W3C trace context headers

## Operations

### Health Checks

#### REST Health Check
- **Endpoint**: `GET /health`
- **Response**: `200 OK` with service status

#### Ready Check
- **Endpoint**: `GET /ready`
- **Response**: `200 OK` when service is ready to accept traffic

### Scaling Guidelines

**Resource Requirements:**
- **CPU:** 0.5-1 core per 1000 RPS
- **Memory:** 512MB base + 100MB per 1000 active instances
- **Storage:** 1GB per 100k process instances

**Recommended Replicas:** 2-3 for production

**Horizontal Scaling:** Stateless design supports horizontal scaling

### Database Operations

#### Running Migrations
```bash
# Automatic (on startup)
go run ./cmd/server

# Manual migration
go run ./internal/migration --path ./db/migration
```

#### Backup Strategy
```bash
# PostgreSQL backup
pg_dump workflow > backup.sql

# Restore
psql workflow < backup.sql
```

#### Connection Pool Settings
- Max Open Connections: 25
- Max Idle Connections: 10
- Connection Max Lifetime: 5 minutes

### Workflow Operations

#### Process Instance Management
```bash
# Get process instances for an entity
curl -H "X-Tenant-ID: DEFAULT" \
  "http://localhost:8081/workflow/v3/transition?entityId=app-123"

# Transition a process instance
curl -X POST -H "X-Tenant-ID: DEFAULT" \
  -H "Content-Type: application/json" \
  -d '{"processId":"proc-123","entityId":"app-123","action":"APPROVE"}' \
  "http://localhost:8081/workflow/v3/transition"
```

#### Escalation Management
```bash
# Trigger auto-escalation
curl -X POST -H "X-Tenant-ID: DEFAULT" \
  -H "Content-Type: application/json" \
  -d '{"attributes":{"role":["ADMIN"]}}' \
  "http://localhost:8081/workflow/v3/auto/APP_REVIEW/_escalate"

# Search escalated applications
curl -H "X-Tenant-ID: DEFAULT" \
  "http://localhost:8081/workflow/v3/auto/_search?processId=proc-123"
```

## Testing

### Running Tests

**All Tests:**
```bash
go test ./...
```

**Unit Tests Only:**
```bash
go test ./internal/...
```

**Integration Tests Only:**
```bash
go test ./tests/...
```

**With Coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**With Verbose Output:**
```bash
go test -v ./...
```

### Test Structure

#### Unit Tests
Located in the same package with `_test.go` suffix:
- `internal/service/process_service_test.go` - Business logic tests
- `internal/repository/postgres/process_repository_test.go` - Database layer tests
- `internal/security/attribute_guard_test.go` - Security layer tests
- `api/handlers/process_handler_test.go` - HTTP handler tests

#### Integration Tests
End-to-end tests in `tests/` directory:
- `tests/integration_test.go` - Complete API flow tests

### Test Dependencies

- **Testify:** `github.com/stretchr/testify` - Assertions and mocks
- **SQLMock:** `github.com/DATA-DOG/go-sqlmock` - Database mocking
- **SQLite:** `github.com/mattn/go-sqlite3` - In-memory database for integration tests

### Mock Setup

```go
// Database mock example
db, mock, err := sqlmock.New()
defer db.Close()

mock.ExpectQuery("SELECT (.+) FROM processes").
    WithArgs(tenantID, processID).
    WillReturnRows(rows)

// Service test
service := services.NewProcessService(repo)
process, err := service.GetProcessByID(ctx, tenantID, processID)
```

## Project Structure

```
workflow/
├── api/                          # API layer
│   ├── handlers/                # HTTP handlers
│   └── routes.go                # Route definitions
├── cmd/server/                  # Application entrypoint
├── config/                      # Configuration management
├── db/migration/                # SQL migration files
├── internal/                    # Private application code
│   ├── migration/               # Migration runner
│   ├── models/                  # Domain models
│   ├── repository/              # Data access layer
│   │   └── postgres/           # PostgreSQL implementations
│   ├── security/               # Security and validation
│   └── service/                # Business logic
├── docker-compose.yml           # Docker compose configuration
├── Dockerfile                   # Docker image definition
├── go.mod                       # Go module definition
└── go.sum                       # Go module checksums
```

## Release & Deployment

### Branching Strategy

**Git Flow:**
- `master` - Production releases
- `develop` - Development integration

### CI/CD Pipeline

TBD

### Versioning

TBD

### Deployment

**Docker Compose (Development):**
```yaml
version: '3.8'
services:
  workflow-go:
    build: .
    ports:
      - "8081:8081"
    environment:
      - DB_HOST=postgres
    depends_on:
      - postgres
  
  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=workflow
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**Kubernetes (Production):**
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workflow-go
spec:
  replicas: 3
  selector:
    matchLabels:
      app: workflow-go
  template:
    metadata:
      labels:
        app: workflow-go
    spec:
      containers:
      - name: workflow-go
        image: workflow-go:latest
        ports:
        - containerPort: 8081
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
```

## Troubleshooting

### Common Issues

#### Database Connection Issues

**Error:** `could not connect to the database`

**Solutions:**
1. Verify PostgreSQL is running
2. Check connection string
3. Verify database exists
4. Check firewall settings

**Debug:**
```bash
# Test database connection
psql -h localhost -U postgres -d workflow
```

#### Migration Issues

**Error:** `migration failed`

**Solutions:**
1. Check migration files syntax
2. Verify database permissions
3. Check for conflicting migrations
4. Review migration logs

**Debug:**
```bash
# Check migration status
psql workflow -c "SELECT * FROM schema_migrations;"
```

#### Guard Validation Failures

**Error:** `guard validation failed`

**Causes:**
- Missing required attributes
- Invalid assignee
- Role mismatch

**Solutions:**
1. Check attribute validation rules
2. Verify user roles and permissions
3. Review assignee configuration
4. Check guard condition logic

### Debug Mode

**Enable Debug Logging:**
```bash
export LOG_LEVEL=debug
go run ./cmd/server
```

**Enable SQL Query Logging:**
```bash
// In configuration
DB_DEBUG=true
```

### Monitoring Queries

**Database Performance:**
```sql
-- Slow queries
SELECT * FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;

-- Connection count
SELECT count(*) FROM pg_stat_activity;

-- Process instance statistics
SELECT status, count(*) FROM process_instances GROUP BY status;
```

### Log Analysis

**Common Log Patterns:**
```bash
# Search for errors
grep "ERROR" application.log

# Find slow requests
grep "duration_ms" application.log | sort -k3 -n

# Analyze by endpoint
grep "/workflow/v3/transition" application.log | head -20
```

## FAQ

### Technical Questions

**Q: How do I create a parallel workflow?**
A: Set `isParallel: true` on the state that should create branches, and `isJoin: true` on the state where branches merge. Use `branchStates` to define the parallel branch state codes.

**Q: How do I configure auto-escalation?**
A: Create escalation configurations for specific process states with SLA thresholds and escalation actions.

**Q: What's the maximum number of states per process?**
A: There's no hard limit, but performance may degrade with very complex workflows (100+ states).

### Operational Questions

**Q: How do I backup the data?**
A: Use PostgreSQL pg_dump for database backup. Process instances contain the complete workflow state.

**Q: Can I run multiple instances?**
A: Yes, the service is stateless and supports horizontal scaling.

**Q: How do I monitor workflow performance?**
A: Use the /health endpoint, Prometheus metrics, and application logs to monitor workflow execution.

## References

TBD

### Support Channels

TBD

---

**Last Updated:** January 2025
**Version:** 3.0.0
**Maintainer:** DIGIT Platform Team
