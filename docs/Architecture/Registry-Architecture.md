# DIGIT Registries

> **Work in Progress**

## Objective

The objective of this document is to define the core principles, architecture, and implementation approach for registries within the DIGIT platform, ensuring they serve as authoritative, interoperable, and scalable data repositories.

By outlining the essential characteristics, design principles, and challenges of government registries, the document aims to guide the development of a federated and decentralized registry framework that enhances data integrity, service efficiency, and secure information exchange across departments.

Additionally, it evaluates different architectural approaches—separate microservices, a common registry service, and a hybrid model—to recommend a scalable and maintainable solution that aligns with government digital transformation objectives.

---

## Introduction

Governments manage vast amounts of critical data across various domains, including citizen identities, properties, businesses, and public services. Registries serve as authoritative repositories that ensure accuracy, reliability, and security, acting as a single source of truth for governance.

However, many government agencies still operate in data silos, leading to duplication, inconsistencies, and inefficiencies in service delivery.

A modern registry system must be:

- **Federated and decentralized**
- **Secure and interoperable**
- **Based on consent and authentication**
- **Standardized in design**
- **Scalable and performance-optimized**

---

## Core Characteristics of a Registry Service

### 1. Be Authoritative: Serve as the Source of Truth

- Trusted and up-to-date information
- Regular updates from verified sources
- Data validation before persistence

**Example:**  
If the Property Registry shows "Plot No. 101" belongs to "John Doe," this must be the single truth across systems.

---

### 2. Support CRUD Operations

- Create, Read, Update, Delete support
- Update operations validated
- Deletion with necessary checks

---

### 3. Provide Search & Query Capabilities

- Full-text search and filtering
- Real-time querying with sorted results

**Example:**  
Water Connection Registry can be queried by address or status (active/inactive).

---

### 4. Uniquely Identifiable

Two types of IDs:
- **Business Unique ID:** Domain-specific and meaningful to users
- **System-Generated UUID:** Ensures global uniqueness internally

**Benefits:**
- Prevents duplicates
- Simplifies referencing
- Enables consistent integration

**Examples:**
- Property ID: `PROP-2025-101`, UUID: `3f50d9b7-e7ba-4e04-a13f-9d3d24b3a5b9`
- National ID or derived identity: `JohnDoe-1990-01-01-CityXYZ`

---

### 5. Traceability

- Audit log captures:
  - Who made changes
  - What was changed
  - When it was changed

**Example:**  
Trade License Registry logs changes in validity dates with metadata.

---

### 6. Interoperability

- Standards-based exchange (JSON, XML)
- Real-time integration (APIs, Webhooks)
- Event-based data synchronization

**Example:**  
Property registry updates trigger tax and water registry syncs.

---

### 7. Scalable & Performant

- Handles large volume and concurrent access
- Uses distributed databases, load balancing, and caching

**Example:**  
Thousands of water connections can be processed during city expansions without performance drops.

---

### 8. Authentication (Record Existence Check)

- Verifies record existence via secure token
- Supports:
  - JWT
  - DID
  - Verifiable Credentials (VC)
  - Passwordless & OTP authentication

**Example:**  
`/registry/authenticate` API returns Yes/No for a property ID.

---

### 9. eKYC (Verified Data Retrieval)

- Consent-based, verified data sharing
- RBAC or ABAC for access control
- Uses Verifiable Credentials and digital signatures

**Example:**  
Banking system fetches KYC from individual registry via `/registry/ekyc/request` and `/registry/ekyc/data`

---

## Key Design Principles for Registries

- **Domain-Driven Design (DDD)** with bounded contexts
- **Data Modeling:** Hierarchical or relational entities
- **Unique Identity Management:** Business IDs and UUIDs
- **Data Integrity:** Validation, transactions
- **Audit Logs:** Immutable history and metadata
- **Security & Access Control:** Role-based, fine-grained
- **Interoperability & Standards:** JSON, XML, CSV, ISO-compliant
- **Scalability:** Sharding, distributed storage
- **APIs & Events:** REST/GraphQL, event-driven architecture
- **Federated & Loosely Coupled:** Independent registries, toggled integrations
- **Configurable**

---

## Key Challenges in Implementing Government Data Registries

### Lack of Unique Identifiers for Records

- National ID inconsistencies (e.g., no national ID in Mozambique)
- Policy restrictions (e.g., birth certificate without ID)

---

### Soft References Instead of Hard References

- Optional references due to missing legacy data

**Example:**  
Water Connection Registry in Punjab lacks land record references.

---

### Non-Validated & Non-Verified Existing Data

- Historical inconsistencies
- Incorrect or duplicate records

**Example:**  
Duplicate business names, outdated land ownership

---

### Cost of Verification for VCs

- Manual verification required for authentic VC issuance
- Expensive and time-consuming

**Example:**  
Engineer must verify physical property details

---

### Interoperability Between Old & New Systems

- Legacy systems on mainframes
- Middleware needed for integration

---

### Security & Privacy Risks

- Data breaches, unauthorized modifications

**Example:**  
Tampering of land ownership records

---

### Lack of Digital Adoption

- Preference for paper-based processes
- Lack of trust in digital systems

---

## Architecture Options

### Approach 1: Separate Microservices for Each Registry

**Pros:**

- Independent deployment & scalability
- Strong domain encapsulation
- Fine-grained access control

**Cons:**

- More complex deployments
- Higher infra costs
- Logic duplication

---

### Approach 2: Common Registry Service

**Pros:**

- Faster development with JSON schema
- Lower infra cost
- Reusable components

**Cons:**

- Weaker domain encapsulation
- Complex validation
- Performance & type safety issues
- Single point of failure
- Complex codebase

---

### Recommendations

| Use Case                        | Recommended Approach         |
| ------------------------------ | ---------------------------- |
| Complex domains with deep logic | Separate Microservices        |
| Simple CRUD registries         | Common Registry Service       |
| Proof of Concept or low-cost   | Common Registry               |
| Mixed workloads                | Hybrid (common + microservices) |

**Hybrid Approach:**
- Common registry for simple domains
- Separate services for business-critical domains
- Shared libraries for logging, security, audit

---
