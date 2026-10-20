# Elydora Go SDK

Official Go SDK for the [Elydora](https://elydora.com) tamper-evident audit platform. Build cryptographically verifiable audit trails for AI agent operations.

## Installation

```bash
go get github.com/Elydora-Infra/Elydora-Go-SDK/v2
```

Requires Go 1.21+. The CLI uses focused TOML and JSONC parsers to preserve user-owned agent configuration.

Upgrading from 2.0.x: run `elydora install` again for every agent. The generated runtimes changed, and `elydora status` does not detect stale runtimes for every adapter.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	elydora "github.com/Elydora-Infra/Elydora-Go-SDK/v2"
)

func main() {
	// Create client with an API token issued in the Console
	client, err := elydora.NewClient(&elydora.Config{
		OrgID:      "org-123",
		AgentID:    "my-agent-id",
		PrivateKey: "<base64url-encoded-ed25519-seed>",
		Token:      "<api-token>",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create and submit an operation
	eor, err := client.CreateOperation(&elydora.CreateOperationParams{
		OperationType: "data.access",
		Subject:       map[string]interface{}{"user_id": "u-123"},
		Action:        map[string]interface{}{"type": "read"},
		Payload:       map[string]interface{}{"record_id": "rec-456"},
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.SubmitOperation(eor)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Receipt: %s\n", resp.Receipt.ReceiptID)
}
```

## CLI

The SDK includes a CLI for installing audit hooks into AI coding agents.

Agent IDs map to one physical directory directly under `~/.elydora`; portable filename rules and physical directory/config checks apply before writes or recursive removal. Ambiguous uninstall discovery requires an explicit agent ID.

```bash
go install github.com/Elydora-Infra/Elydora-Go-SDK/v2/cmd/elydora@latest

elydora install \
  --agent claudecode \
  --org-id org-123 \
  --agent-id agent-456 \
  --private-key-file /secure/path/private.key \
  --token-file /secure/path/api.token \
  --kid agent-456-key-v1
```

Credential options may be omitted in an interactive terminal; the CLI then reads the private key and optional API token without terminal echo. Credential files must contain one UTF-8 line of at most 64 KiB. Unix credential files require owner-only permissions such as `chmod 600`. Provider settings, `guard.js`, `config.json`, `private.key`, and `hook.js` commit as one rollback-capable transaction, and the generated scripts validate protected files at execution time.

Durable managed-file transactions support Windows, Linux, and macOS. The CLI serializes writers across processes, records append-only recovery state under `~/.elydora/transactions`, stages transaction-owned files on each target filesystem, and completes pending recovery before accepting new writes. Windows requires protected owner-only DACLs, hard-link support, atomic `MoveFileEx` replacement, and writable directory handles with `FlushFileBuffers`; Linux filesystems require `renameat2` with `RENAME_NOREPLACE` and `RENAME_EXCHANGE`; macOS filesystems require `renamex_np` with `RENAME_EXCL` and `RENAME_SWAP`. Unsupported platforms and filesystems fail before managed targets change. Recovery and cleanup verify file and directory identities and retain changed or unexpected objects for manual recovery.

Claude Code installation writes exact matchless `PreToolUse`, `PostToolUse`, and `PostToolUseFailure` exec-form hooks to `$CLAUDE_CONFIG_DIR/settings.json` (`~/.claude/settings.json` by default). Elydora preserves unrelated user settings, keeps project, local, managed, plugin, skill, and agent sources unchanged, forwards native snake_case payloads, and uses exit code `2` for frozen or revoked agents. Installation validates the complete official hook schema and commits settings with all four runtime artifacts in one rollback-capable transaction. Run `/hooks` and `claude doctor` after installation to inspect the effective hook sources.

Codex installation writes exact `PreToolUse` and `PostToolUse` command groups to `$CODEX_HOME/hooks.json` (`~/.codex/hooks.json` by default) and preserves the complete native event payload. A configured `CODEX_HOME` follows Codex's existing-directory canonicalization rule. User TOML, project, plugin, and managed sources remain unchanged and continue loading additively. The hook file, generated runtimes, runtime config, and private key commit as one rollback-capable update. Run `/hooks` after installation and approve both Elydora definition hashes.

Cursor installation writes native global `preToolUse`, `postToolUse`, and `postToolUseFailure` handlers to `~/.cursor/hooks.json`. Elydora preserves unrelated user hooks, leaves project and enterprise sources unchanged, emits valid JSON hook responses, fails closed when enforcement or audit execution cannot complete, and commits its 10-second hook contract with all managed runtime files as one rollback-capable transaction.

Gemini CLI installation writes exact matchless `BeforeTool` and `AfterTool` command handlers to `$GEMINI_CLI_HOME/.gemini/settings.json` (`~/.gemini/settings.json` by default). Elydora preserves JSONC comments, line endings, unrelated user hooks, and additive workspace, system, system-override, and extension sources. The user settings document and all four generated runtime artifacts commit in one rollback-capable transaction. Generated hooks retain Gemini's native snake_case payload, emit `{}` on successful or fail-open execution, surface runtime failures through stderr and `error.log`, and use exit code `2` for frozen or revoked agents. Installation honors `hooksConfig` controls and uses encoded PowerShell commands on Windows. Run `/hooks list` after installation to inspect the effective hook inventory.

GitHub Copilot CLI installation writes native `preToolUse`, `postToolUse`, and `postToolUseFailure` user hooks to `$COPILOT_HOME/hooks/elydora-audit.json`, with `~/.copilot/hooks/elydora-audit.json` as the default. Elydora preserves additive hook sources, migrates exact Elydora-owned entries from project `.github/hooks/hooks.json`, validates the official handler schema and JavaScript matcher syntax, resolves `disableAllHooks` through the CLI's settings precedence, and commits all runtime and hook files transactionally. The generated commands preserve native camelCase payloads and the guard propagates exit code `2`; restart active Copilot sessions after installation.

Cline 3.0.46 installation commits `guard.js`, `config.json`, `private.key`, `hook.js`, `PreToolUse.mjs`, and `PostToolUse.mjs` through one rollback-capable transaction. The native hooks live in `$CLINE_DIR/hooks` with `~/.cline/hooks` as the default; the Documents and workspace roots remain read-only. The wrappers forward Cline's complete stdin payload byte-for-byte, use the active Node.js executable for generated runtimes, and emit pure JSON cancellation control for frozen or revoked agents. Status requires physical files, exact generated sources, strict runtime identity, and a canonical private key.

Kiro CLI installation covers both runtime contracts. Kiro CLI v2 uses the generated custom agent through `kiro-cli --agent elydora-audit`. Kiro CLI v3 loads the global standalone hooks when started with `kiro-cli --v3`.

Kiro IDE 1.0 installation writes an exact enabled `PreToolUse` guard and `PostToolUse` audit pair to the active workspace `.kiro/hooks/elydora-audit.json`. Both entries use Kiro's `version: "v1"` schema, `.*` matchers, ten-second timeouts, and platform-safe absolute commands. Elydora preserves unrelated workspace hooks, forwards the complete native event payload, and propagates frozen or revoked state through exit code `2`. Workspace hooks, exact legacy `~/.kiro/hooks/elydora-audit.kiro.hook` migration, generated runtimes, strict runtime metadata, and the canonical private key commit through one rollback-capable transaction. Confirm both entries in Kiro's Agent Hooks panel after installation.

Kimi installation follows home-directory evidence. Kimi Code uses `$KIMI_CODE_HOME/config.toml` with `~/.kimi-code/config.toml` as the default, and legacy `kimi-cli` uses `~/.kimi/config.toml`. Elydora installs exact `PreToolUse`, `PostToolUse`, and `PostToolUseFailure` hooks, preserves native snake_case payloads, and commits all selected TOML documents with the four runtime artifacts in one rollback-capable transaction. Windows commands use encoded PowerShell so paths containing spaces, apostrophes, percent signs, and environment-like text remain literal.

Grok Build installation writes native global `PreToolUse`, `PostToolUse`, and `PostToolUseFailure` hooks to `$GROK_HOME/hooks/elydora-audit.json` (default `~/.grok/hooks/elydora-audit.json`). The hook file, generated runtimes, runtime config, and private key commit as one rollback-capable transaction. Managed commands preserve Grok's native camelCase JSON, return `{"decision":"deny","reason":"..."}` with exit code `2` for frozen or revoked agents, and use encoded PowerShell on Windows. Project hooks follow Grok's `/hooks-trust` workflow; Elydora leaves project, plugin, `hooks-paths`, Claude Code, and Cursor compatibility sources unchanged. Run `grok inspect --json` to inspect the effective hook inventory.

Auggie installation commits `~/.augment/settings.json`, both platform wrappers, generated runtimes, runtime config, and private key as one rollback-capable transaction. The managed `PreToolUse` guard propagates exit code `2`; `PostToolUse` preserves Auggie's complete native hook payload. System, workspace, local workspace, and alternate `--augment-cache-dir` settings remain unchanged. Run `auggie tools list` to validate the effective user configuration.

Factory Droid 0.175 installation follows its whole-source precedence: `~/.factory/hooks.json`, legacy `~/.factory/hooks/hooks.json`, `~/.factory/settings.local.json` `hooks`, then `~/.factory/settings.json` `hooks`. Elydora writes the current `{ "hooks": { ... } }` root contract, preserves JSONC comments and formatting, accepts future hook events, and removes exact legacy Elydora commands during migration. System, project, and nested-folder settings remain read-only policy inputs; `hooksDisabled` and organization-managed hook policy block installation before runtime creation. The guard, native-payload audit runtime, strict config, canonical private key, and every affected hook document commit in one rollback-capable transaction. Windows commands use Factory's native PowerShell execution contract and propagate `$LASTEXITCODE`. Run `/hooks` in Droid after installation to review the effective local and organization-managed hooks.

Qwen Code `0.20.0` installation writes native user hooks to `$QWEN_HOME/settings.json` with `~/.qwen/settings.json` as the default. Elydora follows Qwen's `QWEN_HOME` and `QWEN_RUNTIME_DIR` `.env` bootstrap, reads system defaults, user, trusted workspace, and system override settings in official precedence order, and keeps every non-user source read-only. The adapter preserves JSONC formatting and future event payloads, validates all 21 current events plus command, HTTP, and prompt handlers, applies JavaScript regular-expression semantics only to matcher-enabled events, and owns one named matchless `PreToolUse`, `PostToolUse`, and `PostToolUseFailure` triple. Guard, native-payload audit, strict runtime metadata, canonical private key, and user settings commit in one rollback-capable transaction guarded by source snapshots. Healthy status requires the enabled exact triple and exact generated runtime sources. Run `/hooks` in Qwen Code after installation to review the effective hooks.

### Commands

| Command | Description |
|---------|-------------|
| `elydora install` | Install Elydora audit hook for a coding agent |
| `elydora uninstall` | Remove Elydora audit hook for a coding agent |
| `elydora status` | Show installation status for one agent or all agents |
| `elydora agents` | List supported coding agents |
| `elydora version` | Print the CLI version |

### Supported Agents

| Agent | Key |
|-------|-----|
| Augment Code CLI | `augment` |
| Claude Code | `claudecode` |
| OpenAI Codex | `codex` |
| Cline | `cline` |
| GitHub Copilot CLI | `copilot` |
| Cursor | `cursor` |
| Factory Droid | `droid` |
| Gemini CLI | `gemini` |
| Grok Build | `grok` |
| Kiro CLI | `kirocli` |
| Kiro IDE | `kiroide` |
| Kimi Code | `kimi` |
| Letta Code | `letta` |
| OpenCode | `opencode` |
| Qwen Code | `qwen` |

## API Reference

### Configuration

```go
client, err := elydora.NewClient(&elydora.Config{
	OrgID:      "org-123",       // Organization ID
	AgentID:    "agent-456",     // Agent ID
	PrivateKey: "<seed>",        // Base64url-encoded Ed25519 seed
	BaseURL:    "https://...",   // API base URL (default: https://api.elydora.com)
	TTLMs:      30000,           // Operation TTL in ms (default: 30000)
	MaxRetries: 3,               // Max retries on transient failures (default: 3)
	Token:      "<token>",        // API token for authenticated requests
})
```

### Client Methods

```go
// Update the API token at runtime
client.SetToken("new-api-token")
```

### Authentication

```go
// Get the authenticated user and current organization
me, err := client.GetMe()
// me.User.Email, me.User.Role, me.CurrentOrganization.OrgID

// Issue an API token; this call needs a client whose Token is a Console session token
tokenResp, err := client.IssueApiToken(&elydora.IssueApiTokenRequest{
	TTLSeconds: &ttlSeconds,
})
client.SetToken(tokenResp.Token)

// Rotate the installed API token; the response carries the replacement
rotated, err := client.RotateApiToken()
client.SetToken(rotated.Token)
```

`elydora.Register` and `elydora.Login` remain for password accounts and are deprecated; use the Console sign-in flow.

### Operations

```go
// Create a signed EOR locally (no network call)
eor, err := client.CreateOperation(&elydora.CreateOperationParams{
	OperationType: "inference",
	Subject:       map[string]interface{}{"model": "gpt-4"},
	Action:        map[string]interface{}{"type": "completion"},
	Payload:       map[string]interface{}{"prompt": "Hello"},
	KID:           "agent-456-key-v1",
})

// Submit to API
resp, err := client.SubmitOperation(eor)

// Retrieve an operation
op, err := client.GetOperation(operationID)

// Verify integrity
result, err := client.VerifyOperation(operationID)
```

### Agent Management

```go
// Register a new agent
agent, err := client.RegisterAgent(&elydora.RegisterAgentRequest{
	AgentID:           "my-agent",
	DisplayName:       "My Agent",
	ResponsibleEntity: "team@example.com",
	IntegrationType:   elydora.IntegrationTypeSDK,
	Keys: []elydora.RegisterAgentKeyInput{
		{KID: "key-v1", PublicKey: "<base64url>", Algorithm: "ed25519"},
	},
})

// Get agent details
details, err := client.GetAgent(agentID)

// Change the integration type
updated, err := client.UpdateAgent(agentID, elydora.IntegrationTypeClaudecode)

// Freeze an agent
frozen, err := client.FreezeAgent(agentID, "security review")

// Revoke a key
err := client.RevokeKey(agentID, kid, "key rotation")

// List all agents in the organization
agents, err := client.ListAgents()

// Unfreeze a previously frozen agent
unfrozen, err := client.UnfreezeAgent(agentID, "review complete")

// Delete an agent permanently
deleted, err := client.DeleteAgent(agentID)
```

### Audit

```go
results, err := client.QueryAudit(&elydora.AuditQueryRequest{
	AgentID:       "agent-123",
	OperationType: "inference",
	StartTime:     &startTime,
	EndTime:       &endTime,
	Limit:         &limit,
})
```

### Epochs

```go
epochs, err := client.ListEpochs()
epoch, err := client.GetEpoch(epochID)
```

### Exports

```go
export, err := client.CreateExport(&elydora.CreateExportRequest{
	StartTime: startTime,
	EndTime:   endTime,
	Format:    "json",
})

exports, err := client.ListExports()
detail, err := client.GetExport(exportID)

// Download export file data
data, err := client.DownloadExport(exportID)
```

### Webhooks

```go
webhook, err := client.RegisterWebhook(&elydora.RegisterWebhookRequest{
	EndpointURL: "https://example.com/elydora",
	Events:      []string{"operation.accepted"},
	Secret:      "<signing-secret>",
})
webhooks, err := client.ListWebhooks()
err = client.DeleteWebhook(webhook.Webhook.WebhookID)
```

### Organization

```go
members, err := client.ListMembers()
events, err := client.ListAdminEvents(50)
```

### JWKS

```go
jwks, err := client.GetJWKS()
```

### Health

```go
// Check API health (no authentication required, does not need a client)
health, err := elydora.Health("https://api.elydora.com")
// health.Status, health.Version, health.ProtocolVersion, health.Capabilities, health.Timestamp

// Check dependency health (D1, auth storage, R2, KV)
deep, err := elydora.DeepHealth("https://api.elydora.com")
// deep.Dependencies.D1.Status, deep.Dependencies.KV.LatencyMs
```

### Constants

```go
// Genesis chain hash: initial prev_chain_hash for an agent's first operation
elydora.GenesisChainHash // "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
```

#### Status Enums

| Type | Constants |
|------|-----------|
| `AgentStatus` | `AgentStatusActive`, `AgentStatusFrozen`, `AgentStatusRevoked` |
| `KeyStatus` | `KeyStatusActive`, `KeyStatusRetired`, `KeyStatusRevoked` |
| `ExportStatus` | `ExportStatusQueued`, `ExportStatusRunning`, `ExportStatusDone`, `ExportStatusFailed` |
| `ExportFormat` | `ExportFormatJSON`, `ExportFormatPDF` |
| `WebhookStatus` | `WebhookStatusActive`, `WebhookStatusDisabled` |
| `IntegrationType` | One constant per supported agent key plus `IntegrationTypeEnterprise`, `IntegrationTypeGUI`, `IntegrationTypeSDK`, `IntegrationTypeOther` |
| `AdminAction` | One constant per admin event action, for example `AdminActionAgentFreeze` |
| `RbacRole` | `RbacRoleOrgOwner`, `RbacRoleSecurityAdmin`, `RbacRoleComplianceAuditor`, `RbacRoleReadonlyInvestigator`, `RbacRoleIntegrationEngineer` |
| `ErrorCode` | `ErrorCodeInvalidSignature`, `ErrorCodeUnknownAgent`, `ErrorCodeKeyRevoked`, `ErrorCodeKeyRetired`, `ErrorCodeAgentFrozen`, `ErrorCodeAgentRevoked`, `ErrorCodeTTLExpired`, `ErrorCodeReplayDetected`, `ErrorCodePrevHashMismatch`, `ErrorCodePayloadTooLarge`, `ErrorCodeRateLimited`, `ErrorCodeInternalError`, `ErrorCodeUnauthorized`, `ErrorCodeForbidden`, `ErrorCodeNotFound`, `ErrorCodeValidationError` |

## Error Handling

```go
import "errors"

var apiErr *elydora.ElydoraError
if errors.As(err, &apiErr) {
	fmt.Println(apiErr.Code)       // e.g. "INVALID_SIGNATURE"
	fmt.Println(apiErr.Message)    // Human-readable message
	fmt.Println(apiErr.StatusCode) // HTTP status code
	fmt.Println(apiErr.RequestID)  // Request ID for support
}
```

## License

MIT
