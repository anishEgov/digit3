# Security and Privacy Architecture for Secure, Citizen-Centric Access to Public Services

## Introduction

This document outlines a comprehensive **Security and Privacy Architecture** for citizen access to public services using modern identity and credentialing technologies. It builds on OAuth 2.0, OpenID Connect (OIDC), Decentralized Identifiers (DID), Verifiable Credentials (VCs), and Wallet-based data control. The architecture is designed for **state-level deployment** with centralized identity and federated, department-managed service provisioning.

## 1. Engineering Practices

### Zero Trust Architecture
- Treat all requests as untrusted, regardless of origin
- Enforce authentication and authorization for every access

### Role and Attribute-Based Access Control
- Use Keycloak's role mapping and external attribute injection
- Enforce fine-grained access based on department, user type, service scope

### Principle of Least Privilege
- Clients and tokens are issued with the minimum scopes required
- Each service API validates scopes independently

### Encryption
- **Data in transit**: TLS 1.2+ enforced on all public endpoints
- **Data at rest**: AES-256 encryption for sensitive storage (tokens, secrets, credentials)

### Token Security
- Use short-lived access tokens (e.g., 5–15 mins)
- Implement **refresh token rotation**
- PKCE required for public clients (mobile/web)

### Revocation and Credential Lifecycle
- Verifiable Credentials include revocation support (status lists or registry lookups)
- Revoked credentials are marked in wallet and services are notified

### Auditing and Monitoring
- Log all authentication, token issuance, service access, and credential exchanges
- Include user consent and data delegation in logs
- Enable centralized log aggregation (e.g., ELK or Loki)

## 2. Technology Patterns

### Centralized Identity and Access Management
- Use Keycloak as a central OAuth 2.0 + OIDC provider at the state level
- Departments are clients in Keycloak with dedicated scopes and configurations

### Service-Specific Scope Management
- Department administrators register services and define scopes (e.g., `property.read`, `water.apply`)
- Portal uses Keycloak Admin API to create scopes and bind them to clients

### Self-Service Department Onboarding
- Departments onboard via portal
- Backend provisions Keycloak clients and configures scopes
- Access control and audit rules apply from the point of onboarding

### Secrets and Credential Management
- Use **HashiCorp Vault** to store:
  - Client secrets
  - Signing keys (for VCs)
  - Department API credentials
- Vault is integrated with service deployment pipelines and Keycloak

### Decentralized Credential Verification
- DID-based identifiers for citizens and departments
- VCs issued via OID4VC using OIDC-compatible flows
- Credentials include metadata and proof blocks for verification

### Wallet-Based Credential Control
- Citizens store credentials in standards-based wallets (mobile or browser)
- Wallets support selective disclosure and status checks
- Wallets integrated with login flows and service portals

## 3. Technology Choices

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Authentication & Authorization** | [OAuth 2.0, OpenID Connect](Security/Authentication.md) via Keycloak | Central login, access delegation, token issuance |
| **Credential Issuance & Verification** | [OID4VC, DID, VC](Security/Credentials.md) | Verifiable credentials with revocation support |
| **Credential Wallets** | OpenID4VC-compatible mobile/web wallets | Citizen-managed portable credential storage |
| **Secrets Management** | HashiCorp Vault | Secure storage and access of keys, secrets, API credentials |
| **Audit Logging & Monitoring** | Elastic Stack (ELK), Grafana Loki | Real-time visibility and post-incident analysis |
| **Revocation & Status Registry** | W3C statusList2021 or custom DID registries | Credential revocation and trust validation |

## 4. Technical Flow (High-Level)

1. **Department Onboarding**
   - Admin signs in via central portal
   - Registers department and service offerings
   - Portal creates Keycloak client and assigns scopes

2. **Service and Scope Setup**
   - Admin defines services (e.g., "Property Tax") and associated scopes
   - Portal creates client scopes in Keycloak and binds them to client

3. **Citizen Authentication and Authorization**
   - Citizen logs in via centralized OIDC provider (Keycloak)
   - Approves scopes when accessing specific services
   - ID Token is returned with citizen identity claims

4. **Token Exchange and API Access**
   - Portal exchanges code for access token using PKCE
   - Token used to access APIs with scope validation

5. **VC Issuance**
   - Department issues verifiable credential (e.g., income certificate)
   - Wallet stores the credential securely

6. **VC Presentation and Verification**
   - Another agency requests proof (e.g., welfare benefit check)
   - Wallet presents credential via OID4VC
   - Agency verifies issuer signature and credential status

## 5. Example: Using Keycloak as OAuth Provider

- Keycloak is deployed at `auth.state.gov.in` with a single realm (e.g., `state-services`)
- Each department is registered as a **client** in the realm
- Scopes like `property.read` or `water.apply` are created as **client scopes**
- Admin portal uses **Keycloak Admin API** for dynamic client and scope management
- Keycloak exposes JWKS endpoint for token validation by downstream services
- OIDC login flows return **ID Tokens** with verified citizen identity claims

## 6. Citizen Experience

### Authentication
- Citizen logs in once using centralized portal (`auth.state.gov.in`)
- Accesses all services with SSO (Single Sign-On)

### Verifiable Credentials
- Receives credentials (e.g., income certificate, property ownership) from department
- Stores them in their wallet
- Shares credentials with other departments without re-authentication

### Consent and Control
- Reviews and approves access requests and credential sharing
- Tracks all interactions and shared credentials via wallet

## 7. Summary

| Component | Purpose | Example in Public Services |
|-----------|---------|-----------------------------|
| OAuth 2.0 | Secure access to department APIs | Portal gets permission to access water department API |
| OIDC | Trusted identity from centralized provider | Portal knows citizen's name and ID after login |
| DID | Decentralized identity identifier | Citizen has a portable DID issued by the state |
| VC | Verified, portable digital documents | Income certificate, property ownership, caste proof |
| Wallet | Citizen-controlled storage for data | Mobile app holding all digital credentials |
| OID4VC | Standard protocol for issuing/verifying VCs | Welfare department requests income credential from wallet |
| Vault | Secure key and secret storage | Client secrets and signing keys for departments |
| ELK / Loki | Logging and monitoring | Tracks login, data access, and VC usage |
| StatusList2021 | Revocation infrastructure | Credential verification before acceptance |

By combining centralized identity management with decentralized, citizen-controlled credentials, this architecture ensures secure, privacy-respecting, and scalable access to public services across state-level departments.
