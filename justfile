# github-janitor justfile
# Run `just` to see available recipes

# Default recipe: list available commands
default:
    @just --list

# Build the binary (development)
build:
    go build

# Build the binary (release, stripped)
build-release:
    go build -ldflags="-s -w"

# Run the binary with arguments (use `just -- run <args>` for flags)
run *ARGS:
    go run . {{ARGS}}

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run a specific test by name
test-run NAME:
    go test -run {{NAME}} ./...

# Run tests for a specific package
test-pkg PKG:
    go test ./{{PKG}}

# Format code
fmt:
    go fmt ./...

# Run static analysis
vet:
    go vet ./...

# Run comprehensive linting
lint:
    golangci-lint run

# Run all checks (format, vet, lint, test, tidy, gomod2nix)
check: fmt vet lint test tidy gomod2nix

# Update Go dependencies
update-deps:
    go get -u ./...
    go mod tidy
    gomod2nix > gomod2nix.toml

# Tidy go modules
tidy:
    go mod tidy

# Update gomod2nix.toml after dependency changes
gomod2nix:
    gomod2nix > gomod2nix.toml

# Run govulncheck security scanner
govulncheck:
    govulncheck ./...

# Update flake inputs (package versions)
update:
   nix flake update

# Run with nix
nix-run *ARGS:
    nix run -- {{ARGS}}
