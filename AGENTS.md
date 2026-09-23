# AGENTS.md - AI Agent Guidelines for github-janitor

A Go CLI tool with mise-managed development tasks

**Stack**: Go 1.27+, urfave/cli/v3

## Rules

**Never commit code unless explicitly prompted by the user.**
**Always run `mise run check` after making changes.**

## Commands

Use `mise install` for the Go toolchain and `mise run` for development tasks.

```bash
# Build
mise run build                 # dev build
mise run build-release         # release build

# Run
mise run run -- <args>         # run locally

# Test
mise run test                  # all tests
mise run test-verbose          # verbose test output

# Lint/format
mise run fmt                   # format code
mise run vet                   # static analysis
mise run lint                  # comprehensive linting (golangci-lint)
mise run check                 # run all checks

# Dependencies
mise run tidy                  # go mod tidy
mise run update-deps           # update Go dependencies
```

## Project Structure

```
github-janitor/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # Root command, global flags
│   └── example/                # 'example' subcommand package
│       └── example.go          # Example subcommand
├── internal/
│   ├── cli/
│   │   └── options.go          # GlobalOptions (shared across packages)
│   └── example/
│       └── example.go          # Example internal package
├── go.mod
├── go.sum
└── mise.toml
```

## Code Style

### Imports

Order: stdlib -> external packages -> internal packages, separated by blank lines.
Use goimports or let `go fmt` handle ordering.

### Types

- Use `string` for file paths.
- Use `*T` (pointer) for optional values instead of sentinel values.
- Return `error` for error conditions.

### Naming

- Types: `PascalCase` (MyType, MyStruct)
- Functions/methods: `PascalCase` for exported, `camelCase` for unexported
- Constants: `PascalCase` for exported, `camelCase` for unexported
- Use descriptive names

### Error Handling

- Functions return `(T, error)` tuple.
- Wrap errors with context: `fmt.Errorf("context: %w", err)`.
- Check errors immediately after function calls.
- Use user-facing messages, not debug dumps.

### Formatting

- Run `go fmt ./...` before committing.
- Use `goimports` for import organization.
- Let the tooling handle formatting decisions.

### Testing

- Tests live in `*_test.go` files alongside the code.
- Use table-driven tests where appropriate.
- Use `t.Run` for subtests.
- Prefer exact assertions.

## CLI/UX Guidelines

- `fmt.Println` for normal output, `fmt.Fprintln(os.Stderr, ...)` for errors.
- Avoid breaking existing CLI flags or subcommands.

## Dependency Updates

- Update `go.mod` and run `go mod tidy`.
- Avoid new dependencies unless required.
- Prefer the standard library before adding packages.

## Repo Hygiene

- Keep changes minimal and focused.
- Avoid mass reformatting unless necessary.
