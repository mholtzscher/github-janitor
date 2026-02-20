# Spec: Enable GitHub Security Tooling Across Repositories

## Problem Definition

github-janitor manages repo settings as code, but it cannot yet enable GitHub security tooling in bulk.
Users must click per-repo UI to enable Dependabot-related security features.

This spec adds implementation-ready support for enabling/disabling these security settings across many repositories from `github-janitor.yaml`.

## Scope

### In Scope (v1)

- Manage repository-level security toggles via GitHub API:
  - Dependabot alerts (GitHub "vulnerability alerts")
  - Dependabot security updates (GitHub "automated security fixes")
- Integrate with existing `plan` and `apply` flow.
- Keep current repo selection model: explicit `repositories` list in config.
- Continue-on-error behavior for partial failures (best effort per repo).

### Out of Scope (v1)

- Managing `.github/dependabot.yml`.
- Opening PRs or creating branches.
- Organization-wide repo discovery/query selection.
- GHES support.
- Managing secret scanning/code scanning/advanced security toggles.

## Constraints Inventory

- Target platform: GitHub.com only.
- Must preserve existing CLI UX (`plan`, `apply`, `validate`, `init`).
- Must follow existing pointer-based optional config semantics (`nil` means unmanaged).
- Must continue processing remaining repos if one repo fails.

## Existing Architecture (Discovery)

- Config schema and validation live in `internal/config/config.go`.
- Core reconciliation logic lives in `internal/sync/syncer.go`.
- GitHub API wrapper lives in `internal/github/client.go`.
- `plan` and `apply` both call `SyncAll(ctx, dryRun)` and print `Result.Changes`.
- Custom REST endpoint pattern already exists (`actions/secrets`) using go-github `NewRequest`/`Do`.

## Proposed Design

### Config Schema

Add optional nested `security` settings under `settings`:

```yaml
settings:
  security:
    dependabot_alerts: true
    dependabot_security_updates: true
```

Type model:

```go
type SecuritySettings struct {
    DependabotAlerts          *bool `yaml:"dependabot_alerts,omitempty"`
    DependabotSecurityUpdates *bool `yaml:"dependabot_security_updates,omitempty"`
}

type Settings struct {
    // existing fields...
    Security *SecuritySettings `yaml:"security,omitempty"`
}
```

### Validation Rules

- If `settings.security.dependabot_security_updates == true`, then `settings.security.dependabot_alerts` must be explicitly `true`.
- If `security` is omitted or fields are `nil`, no security changes are managed.

Reason: security updates require dependency graph/vulnerability alerts to be enabled first.

### GitHub API Integration

Use go-github v82 repository helpers via `internal/github/client.go` wrapper methods:

- Vulnerability alerts (Dependabot alerts):
  - Get: `Repositories.GetVulnerabilityAlerts`
  - Enable: `Repositories.EnableVulnerabilityAlerts`
  - Disable: `Repositories.DisableVulnerabilityAlerts`
- Automated security fixes (Dependabot security updates):
  - Get: `Repositories.GetAutomatedSecurityFixes`
  - Enable: `Repositories.EnableAutomatedSecurityFixes`
  - Disable: `Repositories.DisableAutomatedSecurityFixes`

If helper behavior is insufficient, fallback to the existing custom request pattern (`NewRequest`/`Do`) used in the client.

### Sync Semantics

Add security reconciliation per repo in sync engine:

- Read current toggle state(s).
- Compute and append `Result.Changes` entries:
  - `security.dependabot_alerts`
  - `security.dependabot_security_updates`
- Apply changes only when `dryRun == false`.

Mutation ordering:

- Enabling:
  1. Enable `dependabot_alerts` first.
  2. Enable `dependabot_security_updates` second.
- Disabling:
  1. Disable `dependabot_security_updates` first.
  2. Disable `dependabot_alerts` second.

Error behavior:

- Per repo, return wrapped errors with context.
- Continue syncing remaining repos.

## Solution Space + Trade-offs

### Option A: Toggle-only via API (Chosen)

- Pros: minimal scope, no git/PR complexity, fits current architecture.
- Cons: does not manage Dependabot version-update schedules/files.

### Option B: Manage `.github/dependabot.yml` directly

- Pros: full Dependabot configuration control.
- Cons: requires commit/branch/PR semantics, conflict handling, much larger surface.

### Option C: Org policy-level management

- Pros: broad enforcement.
- Cons: different API/domain, permissions model, outside current repo-level model.

## Recommendation

Implement Option A now. Keep schema and sync design extensible for additional security toggles in future versions.

## Deliverables (Ordered)

1. **Config model + validation** (M) - depends on: -
   - Add `SecuritySettings` and `Settings.Security`.
   - Add validation for updates => alerts prerequisite.
   - Update `ExampleConfig()`.

2. **GitHub client security methods** (M) - depends on: 1
   - Add wrapper methods to read/enable/disable both toggles.
   - Normalize API errors with repo + setting context.

3. **Sync integration** (M) - depends on: 2
   - Extend sync interface and add security reconciliation.
   - Ensure dry-run records diffs but performs no mutations.
   - Enforce apply ordering rules.

4. **CLI output integration** (S) - depends on: 3
   - Ensure `plan`/`apply` output prints new fields naturally through existing `Result.Changes` rendering.

5. **Tests** (M) - depends on: 1, 2, 3
   - Config validation tests.
   - Client tests (HTTP method/path/status handling as needed).
   - Sync tests for diffing, ordering, dry-run, and continue-on-error behavior.

6. **Docs update** (S) - depends on: 1
   - README config docs for `settings.security.*`.

## Acceptance Criteria

- `validate` rejects config where `dependabot_security_updates: true` and `dependabot_alerts` is not explicitly `true`.
- `plan` shows per-repo diffs for `security.dependabot_alerts` and `security.dependabot_security_updates` when configured.
- `apply` updates toggles in correct dependency-aware order.
- If one repo fails, remaining repos are still processed and reported.
- Existing non-security settings behavior remains unchanged.
- `just check` passes.

## Risks and Mitigations

- API permission/policy variance across repos (403/404/enterprise policy)
  - Mitigation: per-repo error capture + continue processing + clear field context in errors.
- Ambiguous naming between GitHub terms
  - Mitigation: use stable config keys (`dependabot_alerts`, `dependabot_security_updates`) and document API term mappings.
- API behavior drift or header requirements
  - Mitigation: prefer go-github typed helpers; add focused client tests for status handling.

## Effort Estimate

- Overall: **L** (cross-cutting config + client + sync + tests + docs)

## Open Questions

None.
