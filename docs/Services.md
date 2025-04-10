# DIGIT Platform - Backend API Design Document

## Common Conventions
- **Auth**: All endpoints secured via JWT (issued by Identity Service).
- **Headers**:
  - `Authorization: Bearer <token>`
  - `X-Tenant-ID: <tenant-id>`
- **Format**: JSON over HTTPS
- **Audit Fields** (included in all read responses):
  - `created_by`, `created_on`, `modified_by`, `modified_on`
- **Error Schema**:
```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": "string"
  }
}
```

---

## 1. Account Service

### POST /accounts
```json
{
  "name": "string",
  "domain": "string",
  "oidc_config": { "issuer": "string", "client_id": "string" },
  "administrator": "user_id"
}
```

### GET /accounts/{id}
```json
{
  "id": "string",
  "name": "string",
  "domain": "string",
  "status": "active|closed",
  "administrator": "user_id",
  "oidc_config": { "issuer": "string", "client_id": "string" },
  "created_by": "string",
  "created_on": "datetime",
  "modified_by": "string",
  "modified_on": "datetime"
}
```

### POST /accounts/{id}/status
```json
{ "status": "active|closed" }
```

### POST /accounts/{id}/administrator
```json
{ "administrator": "user_id" }
```

### POST /accounts/{id}/users
```json
{
  "name": "string",
  "email": "string",
  "phone": "string",
  "unique_id": "string",
  "roles": ["string"]
}
```

### GET /accounts/{id}/users
```json
[
  {
    "user_id": "string",
    "name": "string",
    "email": "string",
    "phone": "string",
    "unique_id": "string",
    "roles": ["string"],
    "created_by": "string",
    "created_on": "datetime",
    "modified_by": "string",
    "modified_on": "datetime"
  }
]
```

### POST /accounts/{id}/roles
```json
{ "name": "string", "permissions": ["string"] }
```

### GET /accounts/{id}/roles
```json
[
  { "role_id": "string", "name": "string", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
]
```

---

## 2. Identity Service (OIDC)

### POST /auth/token
```json
{ "username": "string", "password": "string" }
```

### GET /auth/userinfo
```json
{ "sub": "string", "name": "string", "email": "string" }
```

### POST /auth/logout
```json
{ "success": true }
```

### GET /.well-known/openid-configuration
```json
{
  "issuer": "string",
  "authorization_endpoint": "string",
  "token_endpoint": "string",
  "userinfo_endpoint": "string",
  "jwks_uri": "string",
  "response_types_supported": ["string"],
  "subject_types_supported": ["string"],
  "id_token_signing_alg_values_supported": ["string"]
}
```

---

## 3. Catalogue Service

### POST /catalogue/services
```json
{ "name": "string", "description": "string", "endpoint": "string", "configuration": {} }
```

### GET /catalogue/services
```json
[
  {
    "id": "string",
    "name": "string",
    "description": "string",
    "endpoint": "string",
    "configuration": {},
    "created_by": "string",
    "created_on": "datetime",
    "modified_by": "string",
    "modified_on": "datetime"
  }
]
```

### GET /catalogue/services/{id}
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "endpoint": "string",
  "configuration": {},
  "created_by": "string",
  "created_on": "datetime",
  "modified_by": "string",
  "modified_on": "datetime"
}
```

### PUT /catalogue/services/{id}
```json
{ "description": "string", "endpoint": "string", "configuration": {} }
```

---

## 4. Registration Service

### POST /requests
```json
{ "service_id": "string", "data": {} }
```

### GET /requests/{id}
```json
{
  "id": "string",
  "status": "string",
  "data": {},
  "created_by": "string",
  "created_on": "datetime",
  "modified_by": "string",
  "modified_on": "datetime"
}
```

### PUT /requests/{id}
```json
{ "data": {} }
```

### GET /requests?serviceId=&status=&userId=
```json
[
  {
    "id": "string",
    "status": "string",
    "created_by": "string",
    "created_on": "datetime",
    "modified_by": "string",
    "modified_on": "datetime"
  }
]
```

---

## 5. Registry Service

### POST /registry/schemas
```json
{ "name": "string", "schema": {} }
```

### GET /registry/schemas/{name}
```json
{ "name": "string", "schema": {}, "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### POST /registry/data/{schema}
```json
{ "data": {} }
```

### GET /registry/data/{schema}/{id}
```json
{ "id": "string", "data": {}, "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### PUT /registry/data/{schema}/{id}
```json
{ "data": {} }
```

### DELETE /registry/data/{schema}/{id}
```json
{ "deleted": true }
```

### GET /registry/data/{schema}?filters
```json
[
  { "id": "string", "data": {}, "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
]
```

---

## 6. Workflow Service

### POST /workflows
```json
{ "name": "string", "states": [], "transitions": [] }
```

### GET /workflows/{id}
```json
{ "id": "string", "name": "string", "states": [], "transitions": [], "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### POST /workflows/{id}/instances
```json
{ "entity_id": "string" }
```

### POST /workflow-instances/{id}/transition
```json
{ "action": "string" }
```

### GET /workflow-instances/{id}
```json
{ "id": "string", "state": "string", "history": [], "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

---

## 7. Notification Service

### POST /notifications
```json
{ "template_id": "string", "recipient": "string", "data": {} }
```

### GET /notifications/{id}
```json
{ "id": "string", "status": "string", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### POST /templates
```json
{ "name": "string", "content": "string", "type": "email|sms|inapp" }
```

### GET /templates
```json
[
  { "id": "string", "name": "string", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
]
```

---

## 8. File Service

### POST /files
*Request*: Multipart/form-data with file.
```json
{ "id": "string" }
```

### GET /files/{id}
```json
{ "id": "string", "filename": "string", "size": "number", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### GET /files/{id}/url
```json
{ "url": "string" }
```

### DELETE /files/{id}
```json
{ "deleted": true }
```

