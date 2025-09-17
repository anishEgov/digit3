Filestore Service Documentation

## Service Name and Purpose

**Filestore Service**  
A microservice for storing, retrieving, and managing files and their metadata. It provides RESTful APIs for file upload/download, metadata access, and document category management. The service uses MinIO for object storage and PostgreSQL for metadata persistence.

---

## Input/Output Details and Data Formats

### File Upload
- **Endpoint:** `POST /filestore/v1/files/upload`
- **Input:** `multipart/form-data` with fields:
  - `file`: File(s) to upload
  - `tenantId`, `module`, `tag`, `requestInfo` (optional metadata)
- **Output:** JSON with file store IDs

### File Download
- **Endpoint:** `GET /filestore/v1/files/:fileStoreId`
- **Input:** Path parameter `fileStoreId`, query `tenantId`
- **Output:** File stream with appropriate headers

### Metadata Retrieval
- **Endpoint:** `POST /filestore/v1/files/metadata`
- **Input:** JSON with `fileStoreId`, `tenantId`
- **Output:** JSON metadata (file name, type, etc.)

### Get URLs by Tag
- **Endpoint:** `POST /filestore/v1/files/tag`
- **Input:** JSON with `tag`, `tenantId`
- **Output:** JSON list of file URLs

### Get Download URLs
- **Endpoint:** `GET /filestore/v1/files/download-urls`
- **Input:** Query `fileStoreIds`, `tenantId`
- **Output:** JSON mapping fileStoreIds to URLs

### Presigned Upload URL
- **Endpoint:** `POST /filestore/v1/files/upload-url`
- **Input:** JSON with file details
- **Output:** JSON with presigned URL and fileStoreId

### Confirm Upload
- **Endpoint:** `POST /filestore/v1/files/confirm-upload`
- **Input:** JSON with upload confirmation details
- **Output:** JSON confirmation

### Document Category Management
- **Create:** `POST /filestore/v1/files/document-categories`
- **List:** `GET /filestore/v1/files/document-categories`
- **Get by Code:** `GET /filestore/v1/files/document-categories/:docCode`
- **Update:** `PUT /filestore/v1/files/document-categories/:docCode`
- **Delete:** `DELETE /filestore/v1/files/document-categories/:docCode`

---

## List of Endpoints

| Method | Path | Description | Parameters |
|--------|------|-------------|------------|
| GET    | /filestore/health | Health check | - |
| GET    | /filestore/v1/files/:fileStoreId | Download file | Path: fileStoreId, Query: tenantId |
| POST   | /filestore/v1/files/metadata | Get file metadata | JSON: fileStoreId, tenantId |
| POST   | /filestore/v1/files/tag | Get URLs by tag | JSON: tag, tenantId |
| GET    | /filestore/v1/files/download-urls | Get download URLs | Query: fileStoreIds, tenantId |
| POST   | /filestore/v1/files/upload | Upload file(s) | multipart/form-data |
| POST   | /filestore/v1/files/upload-url | Get presigned upload URL | JSON: file details |
| POST   | /filestore/v1/files/confirm-upload | Confirm upload | JSON: confirmation details |
| POST   | /filestore/v1/files/document-categories | Create document category | JSON: category details |
| GET    | /filestore/v1/files/document-categories | List document categories | Query: type (optional) |
| GET    | /filestore/v1/files/document-categories/:docCode | Get category by code | Path: docCode |
| PUT    | /filestore/v1/files/document-categories/:docCode | Update category | Path: docCode, JSON: details |
| DELETE | /filestore/v1/files/document-categories/:docCode | Delete category | Path: docCode |

---

## Dependencies

### Internal
- `gin` (web framework)
- `service/` (Minio, document category, storage services)
- `repository/` (Postgres repository for artifacts)
- `models/` (data models)
- `utils/` (helpers for file handling, env, etc.)

### External
- **MinIO**: S3-compatible object storage
- **PostgreSQL**: Metadata storage
- **GORM**: ORM for Go
- **Gin**: HTTP web framework
- **github.com/minio/minio-go/v7**: MinIO client

---

## Configuration and Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DB_HOST | Database host | minio.default |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | postgres |
| DB_SSL_MODE | Database SSL mode | disable |
| API_ROUTE_PATH | API route prefix | /filestore/v1/files |
| MINIO_ENDPOINT | MinIO endpoint | (none) |
| MINIO_ACCESS_KEY | MinIO access key | (none) |
| MINIO_SECRET_KEY | MinIO secret key | (none) |
| MINIO_BUCKET | MinIO bucket for writes | (none) |
| MINIO_READ_BUCKET | MinIO bucket for reads | (none) |
| MINIO_USE_SSL | Use SSL for MinIO | false |

**Note:** These can be set in the environment or via a `.env` file (see `docker-compose.yml` for mounting).

---

## Code Examples / Usage Snippets

### Upload a File (cURL)
```sh
curl -F "file=@/path/to/file" -F "tenantId=tenant1" http://localhost:8083/filestore/v1/files/upload
```

### Download a File
```sh
curl -O -J "http://localhost:8083/filestore/v1/files/{fileStoreId}?tenantId=tenant1"
```

### Get Presigned Upload URL
```sh
curl -X POST -H "Content-Type: application/json" -d '{"fileName":"test.txt","module":"mod1","tag":"tag1"}' http://localhost:8083/filestore/v1/files/upload-url
```

---

## Known Limitations / TODOs

- **Rate Limiting:** The `/upload-url` endpoint has a TODO for adding rate or capacity limiting.
- **Error Handling:** Some endpoints return generic error messages; more granular error codes may be needed.
- **Bucket Name in Code:** Some helpers (e.g., `GetFolderName`) fetch the bucket name from the environment directly, which may not be consistent with config usage.
- **Sensitive Data:** Ensure secrets (e.g., MinIO keys) are not logged or exposed.

---

## How It Fits Into the Larger System

- **Filestore** is a backend microservice, typically used by other services or frontends to store and retrieve files (documents, images, etc.) in a secure, multi-tenant way.
- It abstracts storage (MinIO) and metadata (Postgres) behind a simple REST API.
- It is designed to be deployed in a containerized environment (see `docker-compose.yml`), and integrates with Kubernetes health checks.

---
