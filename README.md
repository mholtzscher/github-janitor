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

# Initialize configuration file
github-janitor init

# Validate configuration
github-janitor validate

# Preview changes (dry-run)
github-janitor plan

# Apply changes to all repositories
github-janitor sync
```

## Configuration

Create a `.github-janitor.yaml` file to define your repositories and settings:

```yaml
repositories:
  - owner: yourusername
    name: repo1
  - owner: yourusername
    name: repo2

settings:
  # Merge methods
  allow_merge_commit: false
  allow_squash_merge: true
  allow_rebase_merge: true
  delete_branch_on_merge: true

  # Merge commit messages
  squash_merge_commit_title: PR_TITLE      # PR_TITLE, COMMIT_OR_PR_TITLE
  squash_merge_commit_message: PR_BODY     # PR_BODY, COMMIT_MESSAGES, BLANK
  merge_commit_title: PR_TITLE             # PR_TITLE, MERGE_MESSAGE
  merge_commit_message: PR_BODY            # PR_BODY, PR_TITLE, BLANK

  # Repository visibility
  visibility: public                       # public, private

  # Repository features
  has_issues: true
  has_projects: false
  has_wiki: false
  has_discussions: true
  archived: false

  # Additional settings
  allow_update_branch: true
  web_commit_signoff_required: false
  allow_forking: true

  # Repository metadata
  description: "A brief description of the repository"
  homepage: "https://example.com"
  topics: ["go", "cli", "automation"]

  # Repository settings
  default_branch: "main"
  allow_auto_merge: false

  # GitHub Pages (tracks status; enabling requires manual configuration)
  github_pages:
    enabled: false

  # Branch protection
  branch_protection:
    enabled: true
    pattern: "main"
    required_reviews: 1
    dismiss_stale_reviews: true
    require_code_owner_reviews: false
    require_status_checks: true
    require_branches_up_to_date: true
    status_check_contexts: ["ci/test"]
    include_admins: false
    require_linear_history: false
    require_signed_commits: false
    require_conversation_resolution: true
    allow_force_pushes: false
    allow_deletions: false
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
