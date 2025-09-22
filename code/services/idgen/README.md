# ID Generator Service (Go)

A Go-based DIGIT microservice that generates unique, human-readable IDs from user-defined templates. It supports formatted dates, scoped and padded sequences (Postgres-backed for concurrency safety), random segments with flexible charsets, and dynamic variable substitution. Templates are registered via REST APIs, and IDs can be generated at runtime using only the template ID and input data.

## Overview

**Service Name:** idgen

**Purpose:** Provide a robust, configurable, and deterministic ID generation mechanism for DIGIT services using templates composed of literals, variables, date formats, sequences, and random segments.

**Owner/Team:** DIGIT Platform Team

## Architecture

**Tech Stack:**
- Go 1.23
- Gin Web Framework
- PostgreSQL (via GORM)
- Docker

**Core Responsibilities:**
- Register ID generation templates with JSON configuration
- Generate IDs from templates using runtime variables
- Maintain Postgres sequences per template with optional scope-based resets (daily/monthly/yearly/global)
- Validate template correctness (date formats, sequence padding, random charset)
- Run SQL migrations (idempotent, checksum-tracked)

**Dependencies:**
- PostgreSQL 15
- Docker (for containerization)

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

    subgraph "ID Generator Service"
        subgraph "REST API"
            H1[IDGen Handler]
        end

        subgraph "Business Logic"
            S1[IDGen Service]
            P1[Template Parser]
        end

        subgraph "Data Layer"
            R1[IDGen Repository]
        end

        subgraph "Infrastructure"
            M1[Migration Runner]
            CFG[Configuration]
        end
    end

    subgraph "External Systems"
        DB[(PostgreSQL)]
    end

    C1 --> GW
    C2 --> GW
    C3 --> GW

    GW --> H1

    H1 --> S1
    S1 --> P1
    S1 --> R1
    R1 --> DB

    M1 --> DB
```

## Features

- ✅ Template registration with validation
- ✅ Date tokens with multiple keyword formats (e.g., `{DATE:yyyy-mm-dd}`, `{DATE:ddmmyyyy}`)
- ✅ Scoped sequences with custom padding character and length (global/daily/monthly/yearly)
- ✅ Random segment with flexible charsets and range syntax (e.g., `A-Z0-9`)
- ✅ Variable substitution for custom tokens (e.g., `{tenant}`, `{dept}`)
- ✅ Postgres-backed sequences and scope-reset tracking
- ✅ Database migrations with checksum and advisory locks
- ✅ Docker containerization

## Configuration Schema

```json
{
  "template": "{ORG}-{DATE:yyyyMMdd}-{SEQ}-{RAND}",
  "sequence": {
    "scope": "daily",
    "start": 1,
    "padding": { "length": 4, "char": "0" }
  },
  "random": {
    "length": 2,
    "charset": "A-Z0-9"
  }
}
```

### Template
- **Type:** string
- **Value:** A pattern containing static text and dynamic tokens
- **Supported tokens:**
  - `{VARIABLE}`: any client-supplied variable (e.g., `{ORG}`, `{TENANT}`)
  - `{DATE:yyyyMMdd}`: current date formatted via keyword; see Template Syntax for full list
  - `{SEQ}`: a sequence counter, optionally scoped and padded via config
  - `{RAND}`: a random string generated from a configured charset

Example output (for `ORG=PG` on July 3, 2025):

```
PG-20250703-0001-A9
```

### Sequence
Controls behavior of the `{SEQ}` token.

| Property | Type | Description |
|----------|------|-------------|
| `scope` | string | When the counter resets: `daily`, `monthly`, `yearly`, or `global` |
| `start` | integer | Initial value when sequence is first used or after each reset |
| `padding.length` | integer | Minimum width in characters; short values are left-padded |
| `padding.char` | string | Character used for padding (e.g., `"0"` so `1` → `0001`) |

### Random
Controls behavior of the `{RAND}` token.

| Property | Type | Description |
|----------|------|-------------|
| `length` | integer | Number of characters to generate |
| `charset` | string | Allowed characters, supports ranges like `A-Z0-9` |

### Template Syntax

- **Literals:** Any text outside braces, e.g., `INV-`
- **Variables:** `{tenant}`, `{dept}` resolved from request `variables`
- **Date:** `{DATE}` or `{DATE:<keyword>}`; supported keywords include `yyyymmdd`, `ddmmyyyy`, `yyyy-mm-dd`, `dd-mm-yyyy`, `yyyy/mm/dd`, `dd/mm/yyyy`, `yyyy.mm.dd`, `dd.mm.yyyy`, `mmyyyy`, `mm-yyyy`, `yyyy-mm`, `yyyy`, `yy`, etc.
- **Sequence:** `{SEQ}`; uses Postgres sequence `seq_{templateId}` with left padding using `padding.char` repeated to `padding.length`
- **Random:** `{RAND}`; random string of `random.length` from `random.charset` (supports ranges like `A-Z`, `a-z`, `0-9` and combinations like `A-Za-z0-9`)

## Runtime Evaluation Flow

1. Load configuration by `templateId`.
2. Parse the template into an ordered list of segments (static vs tokens).
3. For each segment:
   - Static → append verbatim
   - `DATE` → format `now()` using the requested keyword format
   - `SEQ` → ensure scope reset if needed, fetch `nextval` from Postgres, apply padding
   - `RAND` → pick `length` characters at random from `charset`
   - `{VARIABLE}` → substitute from the provided `variables` map
4. Concatenate all parts to produce the final ID string

## Validation & Error Handling

- Malformed date keyword → validation error when registering the template
- Padding shorter than digits in `start` → validation error when registering the template
- Unknown/empty random charset → validation error when registering the template
- Counter store (Postgres) unavailable → generation blocks/retries until recovery
- Missing required variable for a token → generation error

## Function-style Overview (mapped to REST APIs)

Although exposed as REST, the core responsibilities map cleanly to two functions:

- `registerTemplate(req models.RegisterTemplateRequest): void`
  - REST: `POST /{ctx}/template`
  - Stores the template config in `idgen_templates` and creates a backing Postgres sequence

- `generateId(templateId: string, variables: Map<string,string>): string`
  - REST: `POST /{ctx}/generate`
  - Loads config, evaluates tokens, returns the generated ID

### Storage Model and SQL Examples

Table used for template storage (matching this service's model):

```sql
CREATE TABLE IF NOT EXISTS idgen_templates (
  templateid   character varying(64) PRIMARY KEY,
  config       JSONB NOT NULL,
  createdtime  bigint,
  createdby    character varying(64)
);
```

Per-template Postgres sequence (created on registration):

```sql
-- {templateId} is substituted with the actual template ID
CREATE SEQUENCE IF NOT EXISTS seq_{templateId}
  START WITH {start}
  INCREMENT BY 1
  MINVALUE 1
  CACHE 1;
```

Scope reset tracking (used to ensure sequence reset at boundaries like daily/monthly):

```sql
CREATE TABLE IF NOT EXISTS idgen_sequence_resets (
  templateid character varying(64) NOT NULL,
  scopekey   character varying(32) NOT NULL,
  lastvalue  bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (templateid, scopekey)
);
```

Notes:
- Padding is configured via JSON (`sequence.padding`) rather than inline tokens.
- Date formatting uses predefined keywords; see Template Syntax for supported values.

## Installation & Setup

### Local Development (Manual Setup)

**Prerequisites:**
- Go 1.23+
- PostgreSQL 15

**Steps:**
1. Clone and setup
   ```bash
   git clone https://github.com/digitnxt/digit3.git
   cd code/services/idgen
   go mod download
   ```
2. Setup PostgreSQL database
   ```bash
   createdb idgen_db
   ```
3. Start service (migrations run automatically if enabled)
   ```bash
   go run ./cmd/server
   ```

### Docker

**Build the image:**
```bash
docker build -t idgen:latest .
```

**Run with environment variables:**
```bash
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e MIGRATION_ENABLED=true \
  idgen:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `HTTP_PORT` | Port for REST API server | `8080` | No |
| `SERVER_CONTEXT_PATH` | Base path for API routes | `/idgen` | No |
| `DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `DB_PORT` | PostgreSQL port | `5432` | No |
| `DB_USER` | PostgreSQL username | `postgres` | No |
| `DB_PASSWORD` | PostgreSQL password | `postgres` | Yes |
| `DB_NAME` | PostgreSQL database | `idgen_db` | No |
| `DB_SSL_MODE` | PostgreSQL SSL mode | `disable` | No |
| `MIGRATION_SCRIPT_PATH` | Path to SQL migrations | `./db/migrations` | No |
| `MIGRATION_ENABLED` | Run migrations on startup | `true` | No |
| `MIGRATION_TIMEOUT` | Migration timeout (Go duration) | `5m` | No |

### Example .env file

```bash
HTTP_PORT=8080
SERVER_CONTEXT_PATH=/idgen

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secure_password
DB_NAME=idgen_db
DB_SSL_MODE=disable

MIGRATION_SCRIPT_PATH=./db/migrations
MIGRATION_ENABLED=true
MIGRATION_TIMEOUT=5m
```

## API Reference

Base path is `SERVER_CONTEXT_PATH` (default `/idgen`).

### Templates

#### 1) Register Template
- Endpoint: `POST /{ctx}/template`
- Headers: `X-Client-ID` (optional, recorded as `createdBy`)
- Description: Registers a new ID generation template with configuration and initializes its sequence
- Request Body:
```json
{
  "templateId": "receipt-id",
  "config": {
    "template": "{tenant}-{DATE:yyyy-mm-dd}-{SEQ}-{RAND}",
    "sequence": {
      "scope": "daily",
      "start": 1,
      "padding": { "length": 6, "char": "0" }
    },
    "random": { "length": 4, "charset": "A-Z0-9" }
  }
}
```
- Responses: `201 Created`, `400 Bad Request`, `500 Internal Server Error`

**Sequence Diagram:**
```mermaid
sequenceDiagram
    participant Client
    participant Handler as IDGenHandler
    participant Service as IDGenService
    participant Repo as IDGenRepository
    participant DB as PostgreSQL

    Client->>Handler: POST /{ctx}/template
    Handler->>Service: RegisterTemplate(request)
    Service->>Service: Validate DATE/SEQ/RAND
    Service->>Repo: SaveTemplate()
    Repo->>DB: INSERT idgen_templates
    Service->>Repo: CreateSequence()
    Repo->>DB: CREATE SEQUENCE IF NOT EXISTS seq_{templateId}
    Service-->>Handler: created
    Handler-->>Client: 201 Created
```

### ID Generation

#### 2) Generate ID
- Endpoint: `POST /{ctx}/generate`
- Description: Generates an ID from a registered template using optional variables
- Request Body:
```json
{
  "templateId": "receipt-id",
  "variables": {
    "tenant": "pb",
    "dept": "REV"
  }
}
```
- Success `200 OK`:
```json
{ "id": "pb-2025-09-22-000123-A9XZ" }
```
- Error `400 Bad Request` or `500 Internal Server Error`

**Sequence Diagram:**
```mermaid
sequenceDiagram
    participant Client
    participant Handler as IDGenHandler
    participant Service as IDGenService
    participant Repo as IDGenRepository
    participant DB as PostgreSQL

    Client->>Handler: POST /{ctx}/generate
    Handler->>Service: GenerateID(templateId, variables)
    Service->>Repo: GetTemplate(templateId)
    Repo->>DB: SELECT idgen_templates by templateId
    DB-->>Repo: template JSON
    Service->>Service: parseTemplate(config.template)
    alt token == DATE
        Service->>Service: format time using keyword map
    else token == SEQ
        Service->>Repo: EnsureScopeReset(templateId, scopeKey, start)
        Repo->>DB: INSERT idgen_sequence_resets if new scope
        Service->>Repo: NextSequenceValue(templateId)
        Repo->>DB: SELECT nextval('seq_{templateId}')
    else token == RAND
        Service->>Service: randomString(length, charset)
    else variable token
        Service->>Service: substitute from request variables
    end
    Service-->>Handler: concatenated ID
    Handler-->>Client: 200 OK { id }
```

### Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | BAD_REQUEST | Invalid request parameters/body |
| 500 | INTERNAL_SERVER_ERROR | Server error |
| 500 | GENERATION_FAILED | Template not found/invalid or generation failed |

## Project Structure

```
idgen/
├── cmd/server/                  # Application entrypoint
├── internal/                    # Private application code
│   ├── config/                  # Env config
│   ├── db/                      # Postgres connection
│   ├── handlers/                # HTTP handlers
│   ├── migration/               # Migration runner and tracking
│   ├── models/                  # API & DB models
│   ├── repository/              # Data access layer (templates, sequences)
│   ├── routes/                  # Route definitions
│   └── service/                 # Business logic and template parser
├── db/migrations/               # SQL migration files
├── Dockerfile                   # Docker image definition
├── go.mod                       # Go module definition
└── go.sum                       # Go module checksums
```

## References

TBD

### Support Channels

TBD

---
**Last Updated:** September 2025
**Version:** 1.0.0
**Maintainer:** DIGIT Platform Team
