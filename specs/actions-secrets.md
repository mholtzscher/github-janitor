# Spec: GitHub Actions Secrets Support

## Problem
Enable declarative configuration of GitHub Actions secrets across all configured repositories, with values sourced at runtime from environment variables or executed commands.

## Requirements

### Functional
- Configure secrets via YAML under `settings.actions_secrets`
- Support two sources:
  - `env`: Read value from environment variable
  - `command`: Execute command (argv, no shell), capture stdout as value
- Integrate with existing `plan` and `apply` subcommands
- Upsert semantics (create or update)
- Error on empty/missing source values

### Non-Functional
- Never log or print secret values
- Never execute commands during dry-run/plan mode
- Resolve secrets once, apply to all repos
- Config validation rejects invalid definitions

## Design

### Config Schema

```yaml
settings:
  actions_secrets:
    - name: NPM_TOKEN
      env: NPM_TOKEN
    - name: AWS_ACCESS_KEY_ID
      command: ["op", "read", "op://vault/item/field"]
```

**Validation Rules:**
- `name` required, matches `^[A-Z][A-Z0-9_]*$`, max 64 chars
- Exactly one source: `env` XOR `command`
- `command` must be non-empty array, first element non-empty
- No duplicate `name` entries

### Type Model

```go
// ActionsSecret represents a secret configuration
type ActionsSecret struct {
    Name   string
    Source SecretSource
}

// SecretSource is the interface for value resolution
type SecretSource interface {
    Resolve() (string, error)
    Describe() string  // Safe description for logging (no values)
}

// Concrete implementations:
// - EnvVarSource: env var lookup
// - CommandSource: exec with argv, trim trailing newline
```

### GitHub API Integration

GitHub requires client-side encryption using libsodium sealed boxes:

1. Fetch public key: `GET /repos/{owner}/{repo}/actions/secrets/public-key`
2. Encrypt with `box.SealAnonymous()` (32-byte public key, base64 decoded)
3. Upsert: `PUT /repos/{owner}/{repo}/actions/secrets/{name}`

**Decision:** Direct API implementation vs `gh secret set`:
- Chose direct API: no hard dependency on `gh` CLI, consistent Nix builds, better error control
- Cost: add `golang.org/x/crypto/nacl/box` dependency

### Plan/Apply Behavior

**Dry-Run / Plan:**
- Do NOT resolve sources (no env reads, no command execution)
- Show: `actions_secret.NAME: <unreadable> -> set (source=env:VAR)`

**Apply:**
1. Resolve all secrets once (fail fast on any error)
2. For each repo:
   - Fetch public key
   - Encrypt each secret
   - PUT to GitHub API
3. Never include secret values in results/errors

### Security

- Secrets exist only in memory during apply
- `Describe()` methods safe: `env:NAME` ok, command shows only executable name
- Errors contain source type, never value or full command
- No secret values in `Result.Changes`

## Deliverables (Ordered)

| # | Deliverable | Effort | Depends On |
|---|-------------|--------|------------|
| 1 | Config model + validation + example | M | - |
| 2 | Secret source resolvers (env, command) | M | 1 |
| 3 | GitHub client: Actions secrets REST + encryption | M | - |
| 4 | Sync integration: plan/apply secrets | M | 2, 3 |
| 5 | Tests (config, resolvers, client) | M | 1-4 |
| 6 | Dependency updates + `just check` | S | 1-5 |

## Acceptance Criteria

- [ ] `plan` shows per-repo "would set" entries without touching env/exec
- [ ] `apply` sets secrets across all configured repos
- [ ] Missing/empty env var or command failure aborts before any repo mutations
- [ ] Output never contains secret values or full command argv
- [ ] Config rejects invalid secret names, duplicates, invalid sources
- [ ] `just check` passes

## Open Questions

None for v1.
