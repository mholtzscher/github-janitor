# github-janitor

A Go CLI tool built with Nix

## Installation

### Using Nix

```bash
nix run github:mholtzscher/github-janitor
```

### Using Homebrew

```bash
brew tap mholtzscher/tap
brew install github-janitor
```

### From Source

```bash
git clone https://github.com/mholtzscher/github-janitor.git
cd github-janitor
nix build
```

## Usage

```bash
# Show help
github-janitor --help

# Run example command
github-janitor example

# Run with verbose output
github-janitor --verbose example
```

## Development

This project uses Nix for reproducible development environments.

```bash
# Enter development shell
nix develop

# Or use direnv
direnv allow

# Run checks
just check

# Build
just build

# Run tests
just test
```

## License

MIT
