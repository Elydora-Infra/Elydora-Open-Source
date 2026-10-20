# Contributing to Elydora

Thank you for your interest in contributing to [Elydora](https://elydora.com), the open-source responsibility layer for AI agents. This document explains how to get a development environment running, our code style expectations, and how to submit changes.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Running the API Server](#running-the-api-server)
  - [Running the Console](#running-the-console)
  - [SDK Development](#sdk-development)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Adding a New SDK Language](#adding-a-new-sdk-language)
- [Reporting Issues](#reporting-issues)

---

## Code of Conduct

This project follows the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) Code of Conduct. By participating you agree to abide by its terms. Report unacceptable behavior to the maintainers.

---

## Development Setup

### Prerequisites

| Tool | Minimum version |
|------|----------------|
| Docker | 24+ |
| Docker Compose | v2.20+ |
| Node.js | 18+ (for Console / Node SDK) |
| Python | 3.10+ (for Python SDK) |
| Go | 1.21+ (for Go SDK) |

### Running the API Server

1. **Clone the repository**

   ```bash
   git clone https://github.com/Elydora-Infra/Elydora-Open-Source.git
   cd Elydora-Open-Source
   ```

2. **Configure environment variables**

   ```bash
   cp .env.example .env
   # Edit .env; required variables are documented inside the file
   ```

3. **Generate cryptographic keys**

   ```bash
   ./scripts/generate-keys.sh
   ```

4. **Start all infrastructure services**

   ```bash
   docker compose up -d
   ```

   This starts the API server, PostgreSQL, MinIO (object storage), and Redis.

5. **Verify the server is healthy**

   ```bash
   curl http://localhost:8787/v1/health
   ```

### Running the Console

The web console is a Next.js application located in `packages/console/`.

```bash
cd packages/console
npm install
npm run dev
```

The console is served at `http://localhost:3000`.

### SDK Development

#### Node.js SDK (`sdks/node/`)

```bash
cd sdks/node
npm install
npm run build        # Compile TypeScript to dist/
```

To run the CLI locally without installing:

```bash
node dist/cli.js --help
```

#### Python SDK (`sdks/python/`)

```bash
cd sdks/python
python -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"
```

To run the CLI locally:

```bash
elydora --help
```

Run type checks and tests:

```bash
mypy elydora/
pytest
```

#### Go SDK (`sdks/go/`)

```bash
cd sdks/go
go build ./...
go test ./...
```

To build the CLI binary:

```bash
go build -o elydora ./cmd/elydora
./elydora --help
```

---

## Code Style

### General

- Keep changes focused: one logical change per pull request.
- Write clear commit messages in the imperative mood (`Add X`, `Fix Y`, `Remove Z`).
- Avoid introducing new dependencies without discussion.
- Do not add placeholder data, mock data, or test fixtures to production code paths.

### TypeScript (Node SDK)

- Use strict TypeScript (`"strict": true` in tsconfig).
- Prefer `readonly` properties on interfaces.
- Use explicit return types on exported functions and methods.
- No `any` types; use `unknown` and narrow explicitly.

### Python SDK

- Target Python 3.10+ compatibility.
- Use type hints everywhere (checked with `mypy`).
- Follow PEP 8 with a line length of 100.
- Prefer keyword-only arguments for optional parameters.

### Go SDK

- Follow standard Go formatting (`gofmt`).
- Return `(T, error)`; never panic in library code.
- Keep the public API minimal; unexported helpers start with a lowercase letter.
- Document all exported symbols with Go doc comments.

### API Server

- All route handlers must validate input and return structured `{ error: { code, message, request_id } }` responses on failure.
- All operations must be authorized via the RBAC middleware before handler logic runs.
- Cryptographic operations (signing, verification) must use the canonical JCS representation.

---

## Pull Request Process

1. **Fork** the repository and create a feature branch from `main`.

   ```bash
   git checkout -b feat/your-feature-name
   ```

2. **Make your changes**, keeping commits atomic and well-described.

3. **Test your changes** against a local Docker Compose environment.

4. **Open a pull request** against the `main` branch. Fill in the PR template, including:
   - A clear description of what changed and why.
   - Steps to test the change locally.
   - References to any related issues.

5. **CI checks** run automatically. All checks must pass before a PR can be merged.

6. A maintainer will review your PR. Address any requested changes in new commits (do not force-push to an open PR).

7. Once approved and checks pass, a maintainer will merge using squash-merge.

---

## Adding a New SDK Language

Elydora welcomes official and community SDKs. A conformant SDK must implement the following:

### Required capabilities

| Capability | Details |
|------------|---------|
| EOR construction | Build a signed Elydora Operation Record per the protocol spec |
| Ed25519 signing | Sign the JCS-canonicalized EOR (excluding `signature` field) |
| Chain hashing | SHA-256 of `(prev_chain_hash \|\| payload_hash \|\| operation_id \|\| issued_at)` |
| Operation submission | `POST /v1/operations` |
| Configurable base URL | Must default to `https://api.elydora.com` but accept any URL |
| Retry logic | Retry on 429 and 5xx with exponential backoff |
| Structured errors | Surface `code`, `message`, `request_id` from error responses |

### Recommended capabilities

- Agent registration and management
- Audit query
- Epoch retrieval
- Export creation and download
- Auth helpers (register, login, token issuance)

### Submission

1. Place your SDK under `sdks/<language>/` in the monorepo.
2. Include a `README.md` with installation, quickstart, and API reference.
3. Include a conformance test that exercises at minimum: key generation, EOR construction, signing, and verification against a local server.
4. Open a pull request with the `sdk` label.

---

## Reporting Issues

- Use [GitHub Issues](https://github.com/Elydora-Infra/Elydora-Open-Source/issues) for bug reports and feature requests.
- For security vulnerabilities, **do not open a public issue**. Email the maintainers directly so the issue can be addressed before public disclosure.
- Include as much context as possible: Elydora version, SDK version, OS, steps to reproduce, and expected vs. actual behavior.
