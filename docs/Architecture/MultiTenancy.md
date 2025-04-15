# Multi-Tenancy and Service Management

---

## 1. Overview
DIGIT is a modular, multi-tenant digital public infrastructure designed to streamline public service delivery. It follows a configuration-first design philosophy with extensibility features that empower both business users and developers.

Governments often operate with limited technical capacity, while citizen expectations and service delivery requirements change rapidly. There is a growing need for governments to be able to configure services quickly and independently, without always relying on developer intervention. DIGIT addresses this by enabling rapid service configuration through intuitive tools, while also supporting complex customization where required.

---

## 2. Actor Lifecycle and Responsibilities

### 2.1 Platform Administrator
- Created during initial installation of DIGIT.
- Can add other platform administrators.
- Reviews and approves Service Provider registration requests.
- Manages platform-level:
  - Shared services
  - Reference data
  - Shared registries

### 2.2 Service Provider Administrator
- Gains access upon approval by Platform Administrator.
- Can:
  - Add and manage employees.
  - Configure provider-level:
    - Shared services
    - Reference data
    - Shared registries
  - Set up their own authentication provider (e.g., Google, custom Keycloak).

### 2.3 Service Administrator
- Assigned per service by Service Provider Administrator.
- Can configure:
  - Service-level shared services
  - Reference data
  - Registries
  - Individual service definitions

---

## 3. Service Configuration Workflow

### 3.1 Service Metadata
- Service Name
- Description
- Policy/Process Document
- Type: Configured or Custom Service

### 3.2 Configured Service (via DIGIT Studio)
Service Administrators can configure:
- Forms (UI schemas and layouts)
- Validation rules
- Service charges
- Required documents
- Reference data
- Workflows
- Notifications (email, SMS, in-app)
- PDF output formats
- Certificates (verifiable or standard)
- Extensions (custom validators, logic, hooks)

### 3.3 Custom Service
- External implementation with integration support

---

## 4. Registry Management
Registries can be of two types - Configured or Custom.

### 4.1 Configured Registries
- Defined using:
  - Schema (JSON/YAML)
  - Validation rules
  - Access controls
  - Indexing and constraints
- Can include custom extensions

### 4.2 Custom Registries
- External services or APIs

---

## 5. Extensibility Model
DIGIT supports custom logic at all levels:
- Validator extensions
- Data connectors
- Webhooks and callbacks
- Custom workflow handlers
- Pluggable authentication providers

---

## 6. Value Proposition

| For Business Users                | For Developers                      |
|----------------------------------|-------------------------------------|
| End-to-end service setup         | Inject custom logic as needed       |
| No-code/low-code configuration   | Focus on complex implementations    |
| Empower domain experts           | Maintain extensibility and control  |

---

## 7. Summary
DIGIT balances ease of use with power and flexibility. Business users can fully configure services using DIGIT Studio, while developers can extend functionality through custom implementations. This hybrid approach enables rapid deployment of public services without sacrificing complexity or control.

