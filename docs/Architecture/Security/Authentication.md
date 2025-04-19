# Secure, Citizen-Centric Access to Public Services: From OAuth to Verifiable Credentials

## 1. OAuth: Delegated Access Without Sharing Passwords

### Problem

In public service systems, citizens need to access services from multiple government departments (e.g., property tax, water, health, education). If each system asks them to create a new account or share their password with an aggregator portal, it:

- Increases friction
- Puts the citizen's credentials at risk
- Gives full access instead of limited access

### Solution: OAuth

OAuth allows a citizen to authorize one system (like a central portal) to access data from another system (like the property tax department) without sharing their password.

## OAuth Flow: Step-by-Step

### Step 1: Citizen Clicks "Get My Property Tax Info"

The central portal (client) redirects the citizen to the Property Department's Authorization Server.

```http
GET /authorize
?response_type=code
&client_id=central-portal
&redirect_uri=https://portal.gov.in/callback
&scope=property.read
&code_challenge=abc123
&code_challenge_method=S256
```

- `scope=property.read` means: "only read property tax data."
- `code_challenge` and `code_challenge_method` are part of a security mechanism called Proof Key for Code Exchange (explained below).

### Why Code Challenge Is Needed?

Without it, if someone steals the authorization_code (which is returned in the redirect), they could exchange it for an access token and get the citizen's data.

Proof Key for Code Exchange (PKCE) solves this by:

- Having the portal generate a secret string (code_verifier)
- Hashing it to create code_challenge
- Sending the code_challenge in Step 1
- Later, when requesting the token, the portal proves it has the original code_verifier

### Step 2: Citizen Logs in and Approves Access

They authenticate with the Property Department and approve data sharing.

### Step 3: Property Dept Returns a Code

```
https://portal.gov.in/callback?code=AUTH_CODE_123
```

This is just a temporary code.

### Step 4: Portal Exchanges the Code for an Access Token

```http
POST /token
{
  "grant_type": "authorization_code",
  "code": "AUTH_CODE_123",
  "redirect_uri": "https://portal.gov.in/callback",
  "code_verifier": "abc123"
}
```

### Step 5: Token Issued

```json
{
  "access_token": "eyJhbGciOi...abc",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

The portal now uses this access token to call the Property Dept API.

## 2. OpenID Connect: Understanding Who the Citizen Is

OAuth allows access. But we also need to know: who is the citizen? That's what OpenID Connect (OIDC) adds to OAuth.

### OIDC Adds an ID Token

This is a special token that contains identity information about the citizen.

```json
{
  "id_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsubIjoi12345",
  "name": "Ravi Kumar",
  "email": "ravi@example.com",
  "iss": "https://auth.property.gov.in",
  "aud": "central-portal",
  "exp": 1719999999
}
```

### What Is an ID Token?

An ID token is a JWT (JSON Web Token) — it's a compact, signed string that can be decoded to view structured JSON data.

Fields:
- `sub`: Unique citizen ID
- `iss`: Issuer (who issued this token)
- `aud`: Audience (who this token is for)
- `exp`: Expiry time

This allows the central portal to trust that the citizen is Ravi Kumar, verified by the Property Department.

## 3. Why OAuth and OIDC Are Centralized

In OAuth and OIDC:

- Each department controls its own identity system.
- Portals must ask each department every time data is needed.
- Identity and verification are controlled by centralized servers (e.g., auth.property.gov.in).

This limits:

- Portability (citizen can't carry data across systems)
- Efficiency (need to keep making network calls)
- User control (citizen can't store or share their data independently)

## 4. Decentralized Identifiers (DIDs)

### What Is a DID?

A Decentralized Identifier is a unique ID (like an Aadhaar number or passport number) that is:

- Created and controlled by the citizen
- Stored in a digital wallet
- Looks like: `did:example:123456789abcdef`

### But Can Citizens Just Create Any DID?

Yes, they can. The point is control, not trust. A DID is like a digital address. Anyone can have one, but the value comes when it's linked to verifiable data (see VC below).

Governments can issue government-recognized DIDs like:

```
did:gov:property-dept:45678
```

Citizens can also use their personal DIDs when requesting or presenting data.

## 5. Verifiable Credentials (VCs): Verified, Portable Documents

### Problem

Today, citizens need to submit physical documents for every service:

- Proof of income
- Caste certificate
- Property ownership

These documents must be manually verified, causing delays, duplication, and fraud.

### Solution: Verifiable Credentials

A VC is a digital document signed by an issuing department (like Property or Revenue) that is:

- Pre-verified
- Stored in the citizen's wallet
- Shareable with any department
- Cryptographically signed (tamper-proof)

### Example VC: Property Ownership

```json
{
  "type": ["VerifiableCredential", "PropertyOwnership"],
  "issuer": "did:gov:property-dept",
  "issuanceDate": "2025-04-10",
  "credentialSubject": {
    "id": "did:citizen:xyz123",
    "name": "Ravi Kumar",
    "propertyId": "PROP-1234",
    "owner": true
  },
  "proof": {
    "type": "Ed25519Signature2020",
    "created": "...",
    "verificationMethod": "...",
    "jws": "..."   // digital signature
  }
}
```

### How Does a Verifier Trust This Credential?

The verifier:

1. Checks the signature (using public key of issuer)
2. Looks up the issuer DID (e.g., did:gov:property-dept)
3. Verifies that this DID is trusted (e.g., listed in a government DID registry)

### Challenge: Revocation

If a property is sold, the credential becomes invalid.
Solutions:

- Issuers maintain a revocation list (e.g., status check endpoint or statusList2021 in W3C spec)
- Verifiers check the status before accepting the VC

## 6. Digital Wallets: Citizen Control Over Data

### What Is a Wallet?

A wallet is a secure app (mobile, web, or embedded in a citizen portal) that:

- Stores the citizen's credentials
- Allows them to choose what to share
- Helps them prove ownership of DIDs

Think of it like a physical folder of government documents, but digital and portable.

### Why Wallets Matter

- Citizens are in control (not apps or servers)
- No need to fetch from issuing departments repeatedly
- Enables offline or cross-border access to services

## 7. OpenID for Verifiable Credentials (OID4VC)

### Problem

Current VC issuance and verification flows are not standardized, leading to fragmentation.

### Solution: OID4VC

OID4VC reuses OpenID/OAuth standards to:

- Issue Verifiable Credentials using OIDC flows (OID4VC-I)
- Present Verifiable Credentials using OIDC-based requests (OID4VC-V)

### How It Helps in Public Services

- Agencies can use existing OAuth/OIDC systems to issue credentials
- Verifiers use the same secure protocols to request and validate them
- Citizens use wallets to manage and present credentials

### Example Use Case: Income Certificate

1. Revenue Dept issues an IncomeCredential to citizen's wallet.
2. Later, Social Welfare Dept requests proof of income via OID4VC.
3. Wallet sends the signed credential.
4. Social Welfare verifies it instantly — no need for manual checks.

## Conclusion: A Unified, Secure Public Infrastructure

| Layer | Purpose | Example in Public Services |
|-------|---------|---------------------------|
| OAuth | Access Control | One portal accessing multiple agency APIs |
| OIDC | Identity | Know who the citizen is |
| DID | Decentralized ID | Citizen-managed identity |
| VC | Portable Verified Data | Digital caste, income, property certificates |
| Wallet | Citizen Data Control | Wallet holds all verified credentials |
| OID4VC | Standards for VC exchange | Standard way to issue/verify credentials |

