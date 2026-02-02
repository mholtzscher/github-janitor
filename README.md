# github-janitor

A Go CLI tool built with Nix

## Installation

### Using Nix

```bash
nix run github:mholtzscher/github-janitor
```

### Using Homebrew

```bash
# One-liner
brew install mholtzscher/tap/github-janitor

# Or add the tap explicitly
brew tap mholtzscher/tap
brew install github-janitor

# Upgrade
brew upgrade github-janitor

# Verify
github-janitor --version
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
