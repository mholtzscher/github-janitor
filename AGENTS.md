# AGENTS.md - AI Agent Guidelines for github-janitor

A Go CLI tool built with Nix

**Stack**: Go 1.25+, urfave/cli/v3

## Rules

**Never commit code unless explicitly prompted by the user.**
**Always run `just check` after making changes.**

## Commands

Uses direnv with nix flake for automatic environment setup. Use `just` for development tasks.

```bash
# Build
just build                     # dev build
just build-release             # release build

# Run
just run <args>                # run locally

# Test
just test                      # all tests
just test-verbose              # verbose test output

# Lint/format
just fmt                       # format code
just vet                       # static analysis
just lint                      # comprehensive linting (golangci-lint)
just check                     # run all checks (fmt, vet, lint, test)

# Dependencies
just tidy                      # go mod tidy
just update-deps               # update dependencies and gomod2nix.toml

# Nix build/run
nix build                      # build package
nix run                        # run package
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
└── flake.nix
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
