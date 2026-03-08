# Pennsieve API Gateway Authorization

**Document Owner:** Platform Engineering
**Last Updated:** 2026-03-07
**Classification:** Internal / HIPAA Security Review
**Applicable Regulations:** HIPAA Security Rule (45 CFR 164.312), NIST SP 800-63B

---

## 1. Overview

The Pennsieve API Gateway (`pennsieve-go-api`) is the central entry point for all API requests to the Pennsieve platform. It provides a unified authorization layer that authenticates requests before routing them to downstream microservices (upload-service, datasets-service, packages-service, etc.).

The authorizer supports three distinct authentication flows:

| Flow | Use Case | Credential Type | Entry Point |
|------|----------|-----------------|-------------|
| **Cognito JWT** | Interactive users, API tokens | OAuth 2.0 access token (JWT) | API Gateway HTTP request |
| **Direct Authorization** | Internal service-to-service | Node IDs (no credential) | Lambda-to-Lambda invocation |
| **Callback Token** | Workflow compute containers | Cryptographic bearer token | API Gateway HTTP request |

All three flows produce the same standardized **Claims** output, ensuring downstream services are agnostic to the authentication method used.

---

## 2. Architecture

```
                         ┌─────────────────────────────────┐
                         │       API Gateway (HTTP v2)      │
                         │   upload-service.yml (OpenAPI)   │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                         ┌─────────────────────────────────┐
                         │      Authorizer Lambda          │
                         │                                 │
                         │  ┌───────────────────────────┐  │
                         │  │ Authorization header?      │  │
                         │  │                           │  │
                         │  │ "Bearer ..."  → JWT Flow  │  │
                         │  │ "Callback ..." → Callback │  │
                         │  └───────────────────────────┘  │
                         │                                 │
                         │  Resolves claims via:           │
                         │  • AWS Cognito (JWT validation) │
                         │  • PostgreSQL (user/org/dataset)│
                         │  • DynamoDB (manifest lookup)   │
                         │  • Lambda invoke (callback      │
                         │    token validation)             │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                              Standardized Claims
                         { user, organization, dataset }
                                        │
                                        ▼
                         ┌─────────────────────────────────┐
                         │   Downstream Service Lambda     │
                         │  (upload, datasets, packages…)  │
                         └─────────────────────────────────┘


             ┌─────────────────────────────────┐
             │   Direct Authorizer Lambda      │
             │   (Lambda-to-Lambda only)       │
             │                                 │
             │   Input:  Node IDs              │
             │   Output: Standardized Claims   │
             │   Auth:   AWS IAM (same-account)│
             └─────────────────────────────────┘
```

---

## 3. Flow 1: Cognito JWT Authentication

### 3.1 Description

This is the primary authentication flow for interactive users (web application, CLI) and programmatic API tokens. Clients authenticate with AWS Cognito and present a signed JWT access token.

### 3.2 Request Format

```
Authorization: Bearer <jwt-access-token>
```

### 3.3 Token Sources

The authorizer accepts tokens from two Cognito User Pools:

| Pool | Purpose | Token Lifetime | Client Validation |
|------|---------|---------------|-------------------|
| **User Pool** | Interactive user sessions (web, CLI) | Configured in Cognito (default: 1 hour) | `client_id` must match `USER_CLIENT` |
| **Token Pool** | Programmatic API key access | Configured in Cognito (default: 1 hour) | `client_id` must match `TOKEN_CLIENT` |

### 3.4 Validation Steps

1. **Extract JWT** from `Authorization: Bearer <token>` header
2. **Cryptographic signature verification** using JWKS fetched from Cognito (cached on Lambda cold start)
   - JWKS URL: `https://cognito-idp.{region}.amazonaws.com/{poolId}/.well-known/jwks.json`
   - Keys from both User Pool and Token Pool are combined into a single key set
3. **Issuer validation** — `iss` claim must match one of the two configured pool issuers
4. **Audience validation** — `client_id` claim must match one of the two configured client IDs
5. **Expiration check** — `exp` claim must be in the future
6. **Token use validation** — `token_use` claim must be `"access"`

### 3.5 Claims Resolution

After JWT validation, the authorizer determines the authorization scope based on API Gateway identity sources (query parameters):

| Query Parameter | Authorizer Strategy | Claims Produced |
|-----------------|---------------------|-----------------|
| *(none)* | `UserAuthorizer` | User |
| `dataset_id` | `DatasetAuthorizer` | User + Organization + Dataset |
| `organization_id` | `WorkspaceAuthorizer` | User + Organization + Teams |
| `manifest_id` | `ManifestAuthorizer` | User + Organization + Dataset (via manifest lookup) |

Claims are resolved by querying **PostgreSQL** (via RDS Proxy) for user identity, organization membership, dataset permissions, and team membership. For manifest-based authorization, **DynamoDB** is additionally queried to resolve the manifest's associated dataset.

### 3.6 Security Properties

- **Token integrity**: RSA signature verification using Cognito-managed keys (RS256)
- **Token freshness**: Expiration enforced at authorization time
- **Credential rotation**: JWKS keys are fetched on each Lambda cold start; Cognito handles key rotation
- **Encryption in transit**: All communication over TLS 1.2+
- **No token storage**: The authorizer does not persist tokens; they are validated and discarded

### 3.7 Caching

API Gateway caches authorization results for **300 seconds** (5 minutes). Cache keys are derived from the `identitySource` configuration:
- `$request.header.Authorization` (all routes)
- Plus `$request.querystring.dataset_id`, `$request.querystring.organization_id`, or `$request.querystring.manifest_id` depending on route

This means a token + dataset_id combination is cached separately from the same token + a different dataset_id.

---

## 4. Flow 2: Direct Lambda-to-Lambda Authorization

### 4.1 Description

This flow enables internal Pennsieve services to authorize operations on behalf of a user without requiring a JWT. It is invoked directly via AWS Lambda Invoke (not through API Gateway). The calling service provides node IDs (e.g., `N:user:uuid`, `N:organization:uuid`) and receives resolved claims.

### 4.2 Access Control

- **IAM-scoped**: Only Lambda functions within the same AWS account can invoke the direct authorizer
- **Not externally accessible**: No API Gateway route; not reachable from the internet
- **Trust boundary**: Callers are trusted internal services that have already authenticated the original request through their own mechanisms

### 4.3 Request Format

```json
{
  "user_node_id": "N:user:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "organization_node_id": "N:organization:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "dataset_node_id": "N:dataset:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

- `user_node_id` is required
- `organization_node_id` is optional (required if `dataset_node_id` is provided)
- `dataset_node_id` is optional

### 4.4 Claims Resolution

The direct authorizer queries **PostgreSQL** (via RDS Proxy) to:
1. Look up the user by node ID
2. Resolve organization membership and role
3. Resolve dataset permissions (if dataset_node_id provided)
4. Resolve team memberships

If the user has no access to the requested dataset (`role.None`), authorization is denied.

### 4.5 Security Properties

- **No credential forwarding**: Internal services do not forward user JWTs; they provide verified node IDs
- **AWS IAM enforcement**: Lambda invoke permission is restricted to same-account principals
- **Full permission check**: User's actual permissions are resolved from the database, not assumed or elevated
- **Audit trail**: All invocations are logged via CloudWatch

---

## 5. Flow 3: Callback Token Authentication

### 5.1 Description

This flow enables external compute containers — running workflow analysis pipelines on AWS, Azure, or on-premises infrastructure — to authenticate API requests using a short-lived, run-scoped callback token. This is necessary because these containers do not have access to Cognito user sessions or AWS IAM credentials.

### 5.2 Request Format

```
Authorization: Callback <service-name>:<execution-run-id>:<callback-token>
```

Example:
```
Authorization: Callback workflow-service:550e8400-e29b-41d4-a716-446655440000:a3f1b2c4d5e6f7...
```

The three components are:
- **service-name**: Identifies which service issued the token (e.g., `workflow-service`). Used to route validation to the correct service's validator Lambda.
- **execution-run-id**: The unique identifier of the execution run this token is scoped to.
- **callback-token**: The cryptographic bearer token (64 hex characters = 32 random bytes).

### 5.3 Token Lifecycle

1. **Generation**: When a workflow execution run is created, the workflow-service generates a callback token:
   - 32 cryptographically random bytes via `crypto/rand`
   - SHA-256 hash of the token is stored in DynamoDB (`ExecutionRun.CallbackTokenHash`)
   - Plaintext token is sent to the compute provisioner (never persisted)

2. **Usage**: The compute container includes the token in API requests to Pennsieve endpoints (manifest creation, file upload, status checks).

3. **Expiration**: The token is valid only while the execution run has status `STARTED`. When the run completes, fails, or is cancelled, the token becomes invalid. There is no time-based expiration — token lifetime is tied to run lifecycle.

4. **Revocation**: Changing the run status to any non-`STARTED` value immediately invalidates the token.

### 5.4 Validation Steps

The callback token validation involves two services:

**Step 1: Authorizer Lambda (pennsieve-go-api)**
1. Detect `Callback` prefix in `Authorization` header
2. Parse service name, execution run ID, and token
3. Look up the validator Lambda ARN for the service name from environment configuration
4. If the service name is not registered, deny authorization (401)

**Step 2: Validator Lambda (owned by the issuing service, e.g., workflow-service)**
5. Fetch the `ExecutionRun` record from DynamoDB by execution run ID
6. Verify the run status is `STARTED`
7. Compute SHA-256 hash of the incoming token
8. Compare against the stored `CallbackTokenHash` using **constant-time comparison** (`crypto/subtle.ConstantTimeCompare`) to prevent timing attacks
9. Return the run's context: `userNodeId`, `organizationNodeId`, `datasetNodeId`

**Step 3: Authorizer Lambda (pennsieve-go-api)**
10. Resolve the returned node IDs to full claims via **PostgreSQL** (same resolution as Direct Authorization)
11. Verify the user has access to the dataset (deny if `role.None`)
12. Return standardized claims to API Gateway

### 5.5 Service Registration

Callback token validation is delegated to the service that issued the token. Services are registered via environment variables on the authorizer Lambda:

```
CALLBACK_VALIDATOR_WORKFLOW_SERVICE = arn:aws:lambda:us-east-1:123456789:function:dev-workflow-service-callback-validator-use1
```

Adding a new service requires:
1. The service deploys a validator Lambda implementing the `CallbackValidateRequest`/`CallbackValidateResponse` contract
2. The validator Lambda ARN is added as an environment variable on the authorizer Lambda
3. IAM permission is granted for the authorizer to invoke the validator

No code changes to the authorizer are required to add a new service.

### 5.6 Security Properties

- **Cryptographic strength**: Tokens are 32 bytes of `crypto/rand` output (256 bits of entropy)
- **One-way storage**: Only the SHA-256 hash is stored in DynamoDB; the plaintext token is never persisted at rest
- **Constant-time comparison**: Token validation uses `crypto/subtle.ConstantTimeCompare` to prevent timing side-channel attacks
- **Run-scoped**: Each token is bound to exactly one execution run and its associated user, organization, and dataset
- **Lifecycle-bound**: Tokens are only valid while the execution run is in `STARTED` status; completion or failure immediately invalidates the token
- **Permission-preserving**: The authorizer resolves the original user's actual permissions from the database — the callback token does not grant elevated access beyond what the user who initiated the workflow already has
- **Service isolation**: The authorizer delegates token validation to the issuing service via Lambda-to-Lambda invocation (IAM-scoped, same-account only). The authorizer never accesses the workflow-service's DynamoDB directly.
- **Encryption in transit**: Callback tokens are transmitted over TLS 1.2+ (API Gateway enforces HTTPS). Lambda-to-Lambda invocations use AWS internal TLS.
- **Audit trail**: All callback authorization attempts (success and failure) are logged to CloudWatch with structured JSON including service name and execution run ID. The callback token value is never logged.

### 5.7 Threat Model

| Threat | Mitigation |
|--------|------------|
| Token theft in transit | TLS 1.2+ enforced by API Gateway; tokens never sent over plaintext |
| Token theft from storage | Plaintext token is never stored at rest; only SHA-256 hash in DynamoDB |
| Token reuse after run completion | Validator checks run status is `STARTED`; completed/failed runs reject all tokens |
| Token reuse across datasets | Authorizer resolves claims from the run's specific dataset; token cannot be used for a different dataset |
| Brute-force token guessing | 256 bits of entropy (2^256 possible values); rate limited by API Gateway throttling |
| Timing attack on hash comparison | `crypto/subtle.ConstantTimeCompare` used for all token comparisons |
| Compromised validator Lambda | Validator is IAM-scoped (same-account only); cannot be invoked externally |
| Service impersonation | `X-Callback-Service` value must match a registered service with a configured validator Lambda ARN; unknown services are rejected |
| Privilege escalation | Claims are resolved from the database using the original user's identity; the token cannot grant permissions the user does not have |

### 5.8 Comparison with Cognito JWT

| Property | Cognito JWT | Callback Token |
|----------|-------------|----------------|
| **Token format** | Signed JWT (RS256) | Hex-encoded random bytes |
| **Validation** | Cryptographic signature + expiration | SHA-256 hash comparison + run status |
| **Lifetime** | Time-based (typically 1 hour) | Lifecycle-based (bound to run status) |
| **Revocation** | Not revocable until expiration | Immediate (change run status) |
| **Scope** | User session (multi-resource) | Single execution run (single dataset) |
| **Issued by** | AWS Cognito | Workflow-service (or other registered service) |
| **Requires AWS credentials** | No (standard HTTP header) | No (standard HTTP header) |
| **Suitable for external compute** | No (requires Cognito session) | Yes (works from any environment) |

---

## 6. Standardized Claims Output

All three authorization flows produce the same claims structure, which is passed to downstream services in the API Gateway request context:

```json
{
  "user_claim": {
    "Id": 12345,
    "NodeId": "N:user:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "IsSuperAdmin": false
  },
  "org_claim": {
    "Role": 16,
    "IntId": 67890,
    "NodeId": "N:organization:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "EnabledFeatures": []
  },
  "dataset_claim": {
    "Role": 8,
    "NodeId": "N:dataset:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "IntId": 11111
  },
  "team_claims": [
    {
      "IntId": 1,
      "Name": "publishers",
      "NodeId": "N:team:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      "Permission": 8,
      "TeamType": "publishers"
    }
  ]
}
```

Downstream services use `authorizer.ParseClaims()` from `pennsieve-go-core` to deserialize these claims and `authorizer.HasRole()` to check permissions. **Downstream services are not aware of which authentication flow was used.**

---

## 7. Infrastructure and Network Security

### 7.1 Network Isolation

- Both authorizer Lambdas run in **private VPC subnets** with no internet access
- Database access is via **RDS Proxy** (connection pooling, IAM authentication)
- Lambda-to-Lambda invocations use **AWS internal networking** (no public internet)

### 7.2 Encryption

| Data | At Rest | In Transit |
|------|---------|------------|
| Cognito JWT keys | Managed by AWS Cognito (AWS KMS) | TLS 1.2+ (JWKS fetch) |
| PostgreSQL credentials | IAM-based RDS Proxy auth (no static passwords) | TLS 1.2+ (RDS Proxy) |
| Callback token hashes | DynamoDB server-side encryption (AWS KMS) | TLS 1.2+ (DynamoDB API) |
| API requests | N/A | TLS 1.2+ (API Gateway enforces HTTPS) |

### 7.3 Logging and Monitoring

- All authorization decisions (allow/deny) are logged to **CloudWatch** in structured JSON format
- Logs include: request path, route key, authorization type, service name (for callback), and outcome
- **Sensitive values are never logged**: JWT tokens, callback tokens, database credentials
- API Gateway access logs provide request-level audit trail (source IP, timestamp, status code)
- CloudWatch alarms can be configured for authorization failure rate spikes

---

## 8. Implementation Reference

| Component | Repository | Path |
|-----------|-----------|------|
| API Gateway authorizer (JWT + Callback) | `pennsieve-go-api` | `lambda/authorizer/handler/handler.go` |
| Callback token handler | `pennsieve-go-api` | `lambda/authorizer/handler/callback.go` |
| Direct authorizer | `pennsieve-go-api` | `lambda/authorizer/handler/direct_handler.go` |
| Authorization header parsing | `pennsieve-go-api` | `lambda/authorizer/helpers/helpers.go` |
| Authorizer strategy factory | `pennsieve-go-api` | `lambda/authorizer/factory/factory.go` |
| Claims parsing (downstream) | `pennsieve-go-core` | `pkg/authorizer/claims.go` |
| Callback token generation | `workflow-service` | `internal/compute_trigger/compute_trigger.go` |
| Callback token validator | `workflow-service` | `internal/handler/callback_validator_handler.go` |
| Terraform (authorizer infra) | `pennsieve-go-api` | `terraform/lambda.tf`, `terraform/iam.tf` |
| Terraform (validator infra) | `workflow-service` | `terraform/lambda.tf`, `terraform/outputs.tf` |