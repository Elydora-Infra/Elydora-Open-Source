<div align="center">

<img src="assets/logo.svg" alt="Elydora Logo" width="128" height="128" />

# Elydora

**The Responsibility Layer for AI Agents**

Tamper-evident audit trails with cryptographic proof for every AI agent action.

[![GitHub Stars](https://img.shields.io/github/stars/Elydora-Infra/Elydora-Open-Source?style=social)](https://github.com/Elydora-Infra/Elydora-Open-Source)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Node SDK](https://img.shields.io/badge/node%20sdk%20mirror-v2.1.1-green.svg)](sdks/node/)
[![Python SDK](https://img.shields.io/badge/python%20sdk%20mirror-v2.1.1-green.svg)](sdks/python/)
[![Go SDK](https://img.shields.io/badge/go%20sdk%20mirror-v2.1.1-green.svg)](sdks/go/)
[![CI](https://img.shields.io/github/actions/workflow/status/Elydora-Infra/Elydora-Open-Source/ci.yml?branch=main&label=CI)](https://github.com/Elydora-Infra/Elydora-Open-Source/actions)
[![Website](https://img.shields.io/badge/website-elydora.com-purple.svg)](https://elydora.com)

[Website](https://elydora.com) · [Quick Start](#quick-start) · [Documentation](#documentation) · [API Reference](#api-reference) · [SDKs](#sdks) · [Contributing](#contributing)

---

</div>

## What is Elydora?

Elydora is an open-source **responsibility protocol for AI agents**. When an AI agent reads a file, calls an API, modifies a database, or sends a message, Elydora captures a cryptographically signed, tamper-evident record of that action. These records are chain-hashed together to form an unforgeable audit trail, then rolled up into Merkle tree epochs anchored to public timestamp authorities.

The result is a complete, verifiable history of everything your AI agents have ever done, that no one, including the platform operator, can silently alter.

**Core protocol features:**

- **Ed25519 signed operation records (EOR)**: every agent action is signed with the agent's private key before leaving the agent's process
- **Chain-hashed audit trails**: each record cryptographically commits to the previous record, making silent insertion, deletion, or reordering detectable
- **Merkle tree epoch rollups with TSA anchoring**: periodic epochs hash all records into a Merkle tree whose root is anchored via RFC 3161 Trusted Timestamping
- **Multi-tenant with RBAC**: fine-grained role-based access control with five distinct roles from read-only investigator to organization owner
- **Compliance exports (JSON / PDF)**: one-click export of the complete audit trail for any time range, agent, or operation type

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  AI Agent Process                                               │
│                                                                 │
│  ┌──────────────┐   createOperation()   ┌──────────────────┐  │
│  │  Agent Logic │ ─────────────────────▶│  Elydora SDK     │  │
│  │              │◀─────────────────────  │  (Node/Python/Go)│  │
│  └──────────────┘   submitOperation()   └────────┬─────────┘  │
│                                                   │ Signs EOR   │
│                                                   │ (Ed25519)   │
└───────────────────────────────────────────────────┼─────────────┘
                                                    │ HTTPS
                                                    ▼
┌──────────────────────────────────────────────────────────────────┐
│  Elydora API Server                                              │
│                                                                  │
│  ┌───────────────┐   ┌───────────────┐   ┌──────────────────┐  │
│  │  Auth / RBAC  │   │  Verify EOR   │   │  Epoch Rollups   │  │
│  │  Middleware   │   │  (signature + │   │  (Merkle trees + │  │
│  │               │   │   chain hash) │   │   TSA anchoring) │  │
│  └───────────────┘   └───────────────┘   └──────────────────┘  │
│           │                  │                     │             │
└───────────┼──────────────────┼─────────────────────┼────────────┘
            │                  │                     │
     ┌──────▼──────┐   ┌───────▼────────┐   ┌───────▼──────────┐
     │  PostgreSQL  │   │  Object Store  │   │     Redis        │
     │  (metadata + │   │  (payloads +   │   │  (nonce cache +  │
     │   chain)     │   │   receipts +   │   │   rate limits)   │
     │              │   │   exports)     │   │                  │
     └──────────────┘   └────────────────┘   └──────────────────┘
```

---

## Quick Start

### Prerequisites

- Docker 24+
- Docker Compose v2.20+

### One-Command Deploy

```bash
git clone https://github.com/Elydora-Infra/Elydora-Open-Source.git
cd Elydora-Open-Source

# One-command install: generates all secrets and starts every service
./scripts/install.sh
```

The install script automatically:

1. Checks prerequisites (Docker, Docker Compose, OpenSSL, curl)
2. Generates a `.env` with fresh cryptographic secrets (Ed25519 signing key, Better Auth secret, database and object-store passwords)
3. Starts all services via `docker compose up -d`
4. Waits for the API to become healthy

Once complete, open:

| Service | URL |
|---------|-----|
| Console | http://localhost:3000 |
| API | http://localhost:8787 |
| MinIO Console | http://localhost:9001 |

> **Re-running is safe.** If `.env` already exists, the script skips secret generation and only restarts services.

### Kubernetes Deploy

```bash
helm install elydora ./helm/elydora \
  --namespace elydora \
  --create-namespace \
  -f values.yaml
```

---

## Project Structure

```
Elydora-Open-Source/
├── packages/
│   ├── server/          # Elydora API server
│   └── console/         # Next.js web management console
├── sdks/
│   ├── node/            # @elydora/sdk, Node.js and TypeScript
│   ├── python/          # elydora, sync and async clients
│   └── go/              # github.com/Elydora-Infra/Elydora-Go-SDK/v2
├── helm/
│   └── elydora/         # Kubernetes Helm chart
├── integrations/        # catalog.json: provider IDs and hook contracts
├── scripts/             # Installer, key generation, MinIO bucket setup
├── test/                # Catalog contract tests
├── docker-compose.yml   # One-command self-host stack
├── CONTRIBUTING.md
├── LICENSE
└── README.md
```

| Package | Description |
|---------|-------------|
| `packages/server` | The core API server. Handles EOR ingestion, signature verification, chain validation, RBAC, epoch rollups, and compliance exports. |
| `packages/console` | A web-based management console for viewing agents, browsing the audit trail, managing exports, and administering users. |
| `sdks/node` | TypeScript SDK with full type coverage. Ships a CLI (`elydora`) and a programmatic client. Supports plugin hooks for popular AI coding agents. |
| `sdks/python` | Python SDK with both synchronous (`ElydoraClient`) and asynchronous (`AsyncElydoraClient`) clients. Requires Python 3.10+. |
| `sdks/go` | Go SDK with idiomatic error handling. Ships a CLI binary and embeds zero external runtime dependencies. |
| `helm/elydora` | Production-ready Helm chart for Kubernetes deployments with configurable replicas, resource limits, and secret management. |

---

## Documentation

### Registration and Authentication Flow

```
1. Register organization and first user (Better Auth)
   POST /api/auth/sign-up/email  →  { user, session }

2. Log in to retrieve a session token (Better Auth)
   POST /api/auth/sign-in/email  →  { user, session }

3. Issue a long-lived API token for SDK use
   POST /v1/auth/token  →  { token, expires_at }
   Header: Authorization: Bearer <session-token>

4. Register an agent (as integration_engineer)
   POST /v1/agents/register  →  { agent, keys }

5. Agent submits operations using its SDK client
   POST /v1/operations  →  { receipt (EAR) }
```

### Operation Lifecycle

```
Agent                         SDK                           Server
  │                            │                              │
  │  createOperation(params)   │                              │
  │ ─────────────────────────▶ │                              │
  │                            │  1. Generate UUIDv7 op ID    │
  │                            │  2. Generate 16-byte nonce   │
  │                            │  3. Compute payload_hash     │
  │                            │     SHA-256(JCS(payload))    │
  │                            │  4. Compute chain_hash       │
  │                            │     SHA-256(prev|ph|id|ts)   │
  │                            │  5. Sign JCS(EOR) w/ Ed25519 │
  │                            │  6. Update internal state    │
  │  EOR (signed)              │                              │
  │ ◀───────────────────────── │                              │
  │                            │                              │
  │  submitOperation(eor)      │                              │
  │ ─────────────────────────▶ │  POST /v1/operations         │
  │                            │ ────────────────────────────▶│
  │                            │                              │ Verify signature
  │                            │                              │ Verify chain hash
  │                            │                              │ Check TTL / replay
  │                            │  { receipt: EAR }            │
  │                            │ ◀────────────────────────────│
  │  SubmitOperationResponse   │                              │
  │ ◀───────────────────────── │                              │
```

---

## API Reference

All endpoints are prefixed with the configured base URL (default: `https://api.elydora.com`). All authenticated endpoints require a `Bearer` token in the `Authorization` header.

This self-hosted server mounts `/.well-known/elydora/jwks.json`, `/api/auth/*`, `/v1/auth`, `/v1/agents`, `/v1/operations`, `/v1/audit`, `/v1/epochs`, and `/v1/exports`. The SDKs also expose `/v1/auth/rotate`, `/v1/health/deep`, `/v1/webhooks`, `/v1/members`, and `/v1/admin/events`, which only the hosted API implements; those methods return 404 here. SDK response types follow the hosted API, so fields such as `token_id`, `current_organization`, `onboarding_completed`, `capabilities`, `previous_status`, and the organization's `description` and `ba_org_id` can be absent in self-hosted responses.

Errors are returned as:

```json
{
  "error": {
    "code": "INVALID_SIGNATURE",
    "message": "Ed25519 signature verification failed",
    "request_id": "01J..."
  }
}
```

### Authentication

Authentication is handled by Better Auth. Use the `/api/auth/` endpoints for session management, then issue API tokens for SDK use via `/v1/auth/token`.

#### `POST /api/auth/sign-up/email`

Register a new user. Organization setup is completed during onboarding after registration.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | yes | User email address |
| `password` | string | yes | User password |
| `name` | string | no | Human-readable display name |

**Response:** `{ "user": { ... }, "session": { ... } }`

---

#### `POST /api/auth/sign-in/email`

Authenticate an existing user and receive a session token.

**Request body:** `{ "email": "...", "password": "..." }`

**Response:** `{ "user": { ... }, "session": { "token": "<session-token>", ... } }`

---

#### `GET /api/auth/session`

Return the current session and user profile. Requires a valid session token or cookie.

**Response:** `{ "user": { ... }, "session": { ... } }`

---

#### `POST /v1/auth/token`

Issue an opaque API token for SDK and hook use. Requires a Better Auth session; the token is returned once and only its SHA-256 hash is stored.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ttl_seconds` | integer \| null | no | Token lifetime in seconds, from 1 to 31536000. `null` or omitted for non-expiring. |

**Response:** `{ "token": "<api-token>", "expires_at": <unix seconds> | null }`

---

### Agents

Agents are the AI processes that submit operations. An agent has one or more Ed25519 key pairs registered with the server.

#### `POST /v1/agents/register`

Register a new agent. Required role: **integration_engineer**.

**Request body:**

```json
{
  "agent_id": "my-agent-v1",
  "integration_type": "codex",
  "display_name": "Customer Support Bot",
  "responsible_entity": "engineering-team@example.com",
  "keys": [
    {
      "kid": "my-agent-v1-key-1",
      "public_key": "<base64url Ed25519 public key>",
      "algorithm": "ed25519"
    }
  ]
}
```

**Response:** `{ "agent": { ... }, "keys": [ ... ] }`

---

#### `GET /v1/agents`

List all agents for the organization. Required role: **readonly_investigator**.

**Response:** `{ "agents": [ { "agent_id": "...", "status": "active", ... } ] }`

---

#### `GET /v1/agents/:id`

Retrieve a specific agent and its keys. Required role: **readonly_investigator**.

**Response:** `{ "agent": { ... }, "keys": [ ... ] }`

---

#### `PATCH /v1/agents/:id`

Update agent metadata. Required role: **integration_engineer**.

**Request body:** `{ "integration_type": "codex" }`

---

#### `DELETE /v1/agents/:id`

Permanently delete an agent. Required role: **security_admin**.

**Response:** `{ "deleted": true }`

---

#### `POST /v1/agents/:id/freeze`

Freeze an agent, preventing it from submitting new operations. Required role: **security_admin**.

**Request body:** `{ "reason": "Unusual activity detected" }`

---

#### `POST /v1/agents/:id/unfreeze`

Re-activate a frozen agent. Required role: **security_admin**.

**Request body:** `{ "reason": "Investigation complete" }`

**Response:** `{ "agent": { "status": "active", ... } }`

---

#### `POST /v1/agents/:id/revoke`

Revoke a specific signing key. The agent remains registered but the key can no longer be used to sign new operations. Required role: **security_admin**.

**Request body:** `{ "kid": "my-agent-v1-key-1", "reason": "Key rotation" }`

---

### Operations

#### `POST /v1/operations`

Submit a signed Elydora Operation Record (EOR). Required role: **integration_engineer**. The EOR must be constructed and signed by the agent SDK, never by the server.

**Request body:** A complete EOR (see [Protocol Specification](#protocol-specification)).

**Response:** `{ "receipt": <EAR> }`, the server-issued Elydora Acknowledgment Receipt.

---

#### `GET /v1/operations/:id`

Retrieve a stored operation and its receipt. Required role: **readonly_investigator**.

**Response:** `{ "operation": { ... }, "receipt": { ... } }`

---

#### `POST /v1/operations/:id/verify`

Re-verify a stored operation's cryptographic integrity. Required role: **readonly_investigator**.

**Response:**

```json
{
  "valid": true,
  "checks": {
    "signature": true,
    "chain": true,
    "receipt": true,
    "merkle": true
  },
  "errors": []
}
```

---

### Audit

#### `POST /v1/audit/query`

Query the tamper-evident audit log with filtering and cursor-based pagination. Required role: **compliance_auditor**.

**Request body:**

| Field | Type | Description |
|-------|------|-------------|
| `agent_id` | string | Filter by agent |
| `operation_type` | string | Filter by operation type |
| `start_time` | integer | Start of time range (Unix ms) |
| `end_time` | integer | End of time range (Unix ms) |
| `cursor` | string | Pagination cursor from a previous response |
| `limit` | integer | Maximum number of records to return |

**Response:** `{ "operations": [ ... ], "cursor": "...", "total_count": 1042 }`

---

### Epochs

Epochs are periodic rollups that Merkle-hash all operations in a time window and anchor the root to a public timestamp authority (RFC 3161 TSA).

#### `GET /v1/epochs`

List all epoch records for the organization. Required role: **readonly_investigator**.

**Response:** `{ "epochs": [ { "epoch_id": "...", "root_hash": "...", "leaf_count": 500, ... } ] }`

---

#### `GET /v1/epochs/:id`

Retrieve a specific epoch including its TSA anchor. Required role: **readonly_investigator**.

**Response:**

```json
{
  "epoch": {
    "epoch_id": "...",
    "org_id": "...",
    "start_time": 1700000000000,
    "end_time": 1700003600000,
    "root_hash": "<base64url SHA-256 Merkle root>",
    "leaf_count": 500,
    "created_at": 1700003601234
  },
  "anchor": {
    "tsa_token": "<base64 RFC3161 TimeStampToken>"
  }
}
```

---

### Exports

#### `POST /v1/exports`

Create an asynchronous compliance export job. Required role: **compliance_auditor**.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `start_time` | integer | yes | Start of export window (Unix ms) |
| `end_time` | integer | yes | End of export window (Unix ms) |
| `format` | `"json"` \| `"pdf"` | yes | Export format |
| `agent_id` | string | no | Restrict to one agent |
| `operation_type` | string | no | Restrict to one operation type |

**Response:** `{ "export": { "export_id": "...", "status": "queued", ... } }`

---

#### `GET /v1/exports`

List all exports for the organization. Required role: **compliance_auditor**.

---

#### `GET /v1/exports/:id`

Poll export status and retrieve a download URL once complete. Required role: **compliance_auditor**.

**Response:**

```json
{
  "export": { "export_id": "...", "status": "done", "completed_at": 1700003700000 },
  "download_url": "https://..."
}
```

---

#### `GET /v1/exports/:id/download`

Stream the export file directly. Required role: **compliance_auditor**.

---

### Public Endpoints

#### `GET /v1/health`

Health check. No authentication required.

**Response:**

```json
{
  "status": "ok",
  "version": "1.1.0",
  "protocol_version": "1.0",
  "timestamp": 1700000000000
}
```

---

#### `GET /.well-known/elydora/jwks.json`

The server's JSON Web Key Set. Used by external verifiers to validate EAR signatures. No authentication required.

**Response:** `{ "keys": [ { "kty": "OKP", "crv": "Ed25519", "x": "...", "kid": "...", "use": "sig", "alg": "EdDSA" } ] }`

---

## SDKs

### Node.js / TypeScript

**Installation:**

```bash
npm install @elydora/sdk
```

**Requirements:** Node.js 18+

**Quickstart:**

```typescript
import { ElydoraClient } from '@elydora/sdk';
import { generateKeyPair } from '@elydora/sdk/crypto';

// 1. Register your organization (one-time setup)
const { user, token } = await ElydoraClient.register(
  'http://localhost:8787',
  'admin@example.com',
  'your-password',
  'Admin User',
  'Acme Corp',
);

// 2. Create an agent client
const client = new ElydoraClient({
  orgId: user.org_id,
  agentId: 'order-processor-v1',
  privateKey: process.env.AGENT_PRIVATE_KEY!,
  baseUrl: 'http://localhost:8787', // omit for https://api.elydora.com
});
client.setToken(token);

// 3. Register the agent with its public key
await client.registerAgent({
  agent_id: 'order-processor-v1',
  integration_type: 'sdk',
  display_name: 'Order Processor',
  keys: [{ kid: 'order-processor-v1-key-1', public_key: client.getPublicKey(), algorithm: 'ed25519' }],
});

// 4. Submit an operation
const eor = client.createOperation({
  operationType: 'order.process',
  subject: { order_id: 'ord-123', customer_id: 'cust-456' },
  action: { verb: 'charge', amount_cents: 4999, currency: 'USD' },
});
const { receipt } = await client.submitOperation(eor);
console.log('Accepted. seq_no:', receipt.seq_no);
```

**CLI:**

```bash
npx elydora health --base-url http://localhost:8787
npx elydora keygen               # Generate a new Ed25519 key pair
npx elydora submit-op ...        # Submit an operation from the CLI
```

---

### Python

**Installation:**

```bash
pip install elydora
```

**Requirements:** Python 3.10+, `requests`, `aiohttp`, `cryptography`

**Quickstart (synchronous):**

```python
from elydora import ElydoraClient

# One-time registration
resp = ElydoraClient.register(
    "http://localhost:8787",
    "admin@example.com",
    "your-password",
    display_name="Admin User",
    org_name="Acme Corp",
)

# Create agent client
client = ElydoraClient(
    org_id=resp["user"]["org_id"],
    agent_id="order-processor-v1",
    private_key=os.environ["AGENT_PRIVATE_KEY"],
    base_url="http://localhost:8787",  # omit for https://api.elydora.com
    token=resp["token"],
)

# Register the agent and submit operations
from elydora.crypto import derive_public_key
client.register_agent({
    "agent_id": "order-processor-v1",
    "integration_type": "sdk",
    "display_name": "Order Processor",
    "keys": [{"kid": "order-processor-v1-key-1",
               "public_key": derive_public_key(os.environ["AGENT_PRIVATE_KEY"]),
               "algorithm": "ed25519"}],
})

eor = client.create_operation(
    operation_type="order.process",
    subject={"order_id": "ord-123", "customer_id": "cust-456"},
    action={"verb": "charge", "amount_cents": 4999, "currency": "USD"},
)
receipt = client.submit_operation(eor)
print("Accepted. seq_no:", receipt["receipt"]["seq_no"])
```

**Quickstart (asynchronous):**

```python
from elydora import AsyncElydoraClient

async def main():
    client = AsyncElydoraClient(
        org_id="...", agent_id="...", private_key="...",
        base_url="http://localhost:8787",
    )
    eor = client.create_operation(operation_type="file.read", subject={...}, action={...})
    receipt = await client.submit_operation(eor)
```

**CLI:**

```bash
elydora health --base-url http://localhost:8787
elydora keygen
```

---

### Go

**Installation:**

```bash
go get github.com/Elydora-Infra/Elydora-Go-SDK/v2
```

**Requirements:** Go 1.21+

**Quickstart:**

```go
package main

import (
    "fmt"
    "os"

    elydora "github.com/Elydora-Infra/Elydora-Go-SDK/v2"
)

func main() {
    // One-time registration
    reg, err := elydora.Register(
        "http://localhost:8787",
        "admin@example.com",
        "your-password",
        elydora.WithDisplayName("Admin User"),
        elydora.WithOrgName("Acme Corp"),
    )
    if err != nil {
        panic(err)
    }

    // Create agent client
    client, err := elydora.NewClient(&elydora.Config{
        OrgID:      reg.User.OrgID,
        AgentID:    "order-processor-v1",
        PrivateKey: os.Getenv("AGENT_PRIVATE_KEY"),
        BaseURL:    "http://localhost:8787", // omit for https://api.elydora.com
        Token:      reg.Token,
    })
    if err != nil {
        panic(err)
    }

    // Submit an operation
    eor, err := client.CreateOperation(&elydora.CreateOperationParams{
        OperationType: "order.process",
        Subject:       map[string]interface{}{"order_id": "ord-123"},
        Action:        map[string]interface{}{"verb": "charge", "amount_cents": 4999},
    })
    if err != nil {
        panic(err)
    }

    result, err := client.SubmitOperation(eor)
    if err != nil {
        panic(err)
    }
    fmt.Println("Accepted. seq_no:", result.Receipt.SeqNo)
}
```

**CLI:**

```bash
go run ./cmd/elydora health --base-url http://localhost:8787
go run ./cmd/elydora keygen
```

---

## Protocol Specification

### EOR: Elydora Operation Record

The EOR is the fundamental unit of the protocol. It is constructed and signed by the agent SDK before being sent to the server. The server never modifies an EOR after signing.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `op_version` | `"1.0"` | Protocol version |
| `operation_id` | string | UUIDv7, monotonic and time-ordered |
| `org_id` | string | Organization identifier |
| `agent_id` | string | Agent identifier |
| `issued_at` | integer | Unix timestamp in milliseconds when the EOR was constructed |
| `ttl_ms` | integer | Validity window in milliseconds (default: 30,000) |
| `nonce` | string | 16 random bytes, base64url-encoded, prevents replay |
| `operation_type` | string | Application-defined operation category (e.g. `"file.read"`, `"order.process"`) |
| `subject` | object | The entity the operation acts upon |
| `action` | object | Description of what was done |
| `payload` | object \| string \| null | Optional structured payload |
| `payload_hash` | string | SHA-256 of JCS-canonicalized `payload`, base64url |
| `prev_chain_hash` | string | Chain hash of the immediately preceding EOR for this agent |
| `agent_pubkey_kid` | string | Key ID of the signing key |
| `signature` | string | Ed25519 signature over JCS-canonical EOR (excluding `signature` field), base64url |

**Signing:**

```
canonical_bytes = JCS_canonicalize(EOR without "signature" field)
signature = Ed25519_sign(agent_private_key, canonical_bytes)
```

**Chain hashing:**

```
chain_hash = SHA-256(
  prev_chain_hash + "|" + payload_hash + "|" + operation_id + "|" + issued_at
)
```

All inputs are UTF-8 strings joined with `|`. The genesis (first) `prev_chain_hash` for every agent is the base64url encoding of 32 zero bytes: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=`.

**Payload hashing:**

```
payload_hash = SHA-256(JCS_canonicalize(payload))
```

If `payload` is `null`, the hash is of the JCS canonical form of `null`.

---

### EAR: Elydora Acknowledgment Receipt

The EAR is issued by the server upon accepting a valid EOR. It constitutes server-side proof of ingestion.

| Field | Type | Description |
|-------|------|-------------|
| `receipt_version` | string | Receipt protocol version |
| `receipt_id` | string | Unique receipt identifier |
| `operation_id` | string | Echo of the operation ID |
| `org_id` | string | Organization ID |
| `agent_id` | string | Agent ID |
| `server_received_at` | integer | Unix ms when the server accepted the EOR |
| `seq_no` | integer | Monotonically increasing sequence number for this agent |
| `chain_hash` | string | Server-computed chain hash confirming the EOR was inserted in the correct position |
| `queue_message_id` | string | Internal durable queue message identifier |
| `receipt_hash` | string | SHA-256 of the canonical receipt content |
| `elydora_kid` | string | Key ID of the server's signing key |
| `elydora_signature` | string | Server's Ed25519 signature over the receipt |

---

### EER: Elydora Epoch Record

Epochs are created periodically. All EOR chain hashes within the epoch window are hashed into a binary Merkle tree. The root hash is then anchored to an RFC 3161-compliant Trusted Timestamp Authority.

| Field | Type | Description |
|-------|------|-------------|
| `epoch_id` | string | Unique epoch identifier |
| `org_id` | string | Organization ID |
| `start_time` | integer | Epoch start (Unix ms) |
| `end_time` | integer | Epoch end (Unix ms) |
| `root_hash` | string | Merkle root of all chain hashes in the window |
| `leaf_count` | integer | Number of operations included |
| `created_at` | integer | When the epoch record was created |

The TSA token is an RFC 3161 `TimeStampToken` proving the Merkle root existed at or before the timestamp authority's stated time.

---

## RBAC Roles

| Role | Level | Capabilities |
|------|-------|-------------|
| `org_owner` | 50 | Full administrative access; can manage all users, roles, agents, and organization settings |
| `security_admin` | 40 | Agent lifecycle management: freeze, unfreeze, revoke keys, delete agents |
| `compliance_auditor` | 30 | Read audit log, query operations, create and download compliance exports |
| `integration_engineer` | 20 | Register agents, submit operations, update agent metadata |
| `readonly_investigator` | 10 | Read-only access to agents, operations, epochs; cannot modify any state |

Higher-level roles inherit all capabilities of lower-level roles.

---

## Configuration

Elydora is configured via environment variables. Copy `.env.example` to `.env` and configure the following:

### Required

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `STORAGE_ENDPOINT` | Object storage endpoint (S3-compatible, e.g. MinIO) |
| `STORAGE_BUCKET` | Bucket name for payloads, receipts, and exports |
| `STORAGE_ACCESS_KEY_ID` | Object storage access key |
| `STORAGE_SECRET_ACCESS_KEY` | Object storage secret key |
| `BETTER_AUTH_SECRET` | Better Auth secret for session management |
| `BETTER_AUTH_URL` | Better Auth base URL (e.g. `http://localhost:8787`) |
| `SERVER_PRIVATE_KEY` | Base64url-encoded Ed25519 seed for signing EARs and JWKS |

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8787` | HTTP port the API server listens on |
| `EPOCH_INTERVAL_SECONDS` | `3600` | How often epoch rollups are created |
| `TSA_URL` | none | RFC 3161 Trusted Timestamp Authority URL. If unset, TSA anchoring is disabled. |
| `LOG_LEVEL` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |
| `MAX_PAYLOAD_BYTES` | `65536` | Maximum size of an EOR payload |
| `OPERATION_TTL_MAX_MS` | `300000` | Maximum allowable `ttl_ms` in submitted EORs |
| `CORS_ORIGINS` | `*` | Allowed CORS origins for the console |

---

## Security

### Cryptographic Guarantees

**Ed25519 signatures.** Every EOR is signed by the agent's Ed25519 private key before leaving the agent's process. The server verifies this signature against the registered public key before accepting the record. A compromised server cannot forge valid EORs without access to the agent's private key.

**SHA-256 chain hashing.** Each EOR commits to the hash of the previous EOR, the payload, the operation ID, and the timestamp. Silent insertion, deletion, or reordering of records in the chain is detectable because any change invalidates all subsequent chain hashes.

**PBKDF2-SHA256 password hashing.** User passwords are hashed with PBKDF2-SHA256 before storage.

**Better Auth session-based authentication.** User sessions are managed by Better Auth with secure server-side session storage. Session tokens cannot be forged without the secret.

**RFC 3161 TSA anchoring.** Epoch Merkle roots are submitted to a public timestamp authority, providing independent third-party attestation of when records existed. This prevents retroactive alteration of historical data even by the platform operator.

**JCS canonicalization (RFC 8785).** All signing operations use JSON Canonicalization Scheme, eliminating ambiguity in JSON serialization and ensuring signatures are deterministic across all SDK implementations.

### Key Management

- Agent private keys never leave the agent process. The SDK generates signatures locally.
- The server holds only the agent's public key.
- Server signing keys for EARs and JWKS are stored as environment variables and rotated via `kid` versioning.
- Key rotation is supported: register additional keys with the same agent and revoke old keys without disrupting the audit trail.

### Replay Protection

- Each EOR includes a cryptographically random 16-byte nonce.
- The server maintains a nonce cache in Redis to detect and reject replayed operations within the TTL window.
- The `ttl_ms` field defines the validity window; operations submitted after `issued_at + ttl_ms` are rejected.

---

## Star This Repo

If Elydora is useful to you, please consider giving it a star on GitHub. It helps others discover the project and motivates continued development.

[![Star on GitHub](https://img.shields.io/github/stars/Elydora-Infra/Elydora-Open-Source?style=social)](https://github.com/Elydora-Infra/Elydora-Open-Source)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code style guidelines, and the pull request process.

Contributions are welcome across all parts of the project: server, console, SDKs, Helm chart, documentation, and new SDK language implementations.

---

## License

Copyright 2025-present Elydora Contributors

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full text.
