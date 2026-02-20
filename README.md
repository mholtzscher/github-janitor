# github-janitor

A CLI tool for managing GitHub repository settings as code. 

Instead of clicking through the GitHub UI to configure branch protections, merge methods, and repo features for your projects, you can define them in a `github-janitor.yaml` file and apply them across multiple repositories at once.

## Installation

**Nix**
```bash
nix run github:mholtzscher/github-janitor
```

**Homebrew**
```bash
brew install mholtzscher/tap/github-janitor
```

**From source**
```bash
git clone https://github.com/mholtzscher/github-janitor.git
cd github-janitor
nix build
```

## Authentication

The tool needs GitHub API access. It will automatically use your credentials if you're logged in with the GitHub CLI (`gh auth login`). Alternatively, you can provide a personal access token via the `GITHUB_TOKEN` environment variable.

## Usage

Generate a default configuration file in your current directory:
```bash
github-janitor init
```

Validate the config syntax:
```bash
github-janitor validate
```

Preview what changes will be made (dry run):
```bash
github-janitor plan
```

Apply the settings to all configured repositories:
```bash
github-janitor apply
```

## Configuration

Your `github-janitor.yaml` file defines both the target repositories and the settings you want to enforce. 

```yaml
repositories:
  - owner: yourusername
    name: repo1
  - owner: yourusername
    name: repo2

settings:
  description: "A brief description of the repository"
  homepage: "https://example.com"
  topics: ["go", "cli", "automation"]
  visibility: public
  default_branch: "main"
  archived: false

  # Features
  has_issues: true
  has_projects: false
  has_wiki: false
  has_discussions: true

  # Merge Settings
  allow_merge_commit: false
  allow_squash_merge: true
  allow_rebase_merge: true
  delete_branch_on_merge: true
  allow_auto_merge: false

  squash_merge_commit_title: PR_TITLE      # PR_TITLE, COMMIT_OR_PR_TITLE
  squash_merge_commit_message: PR_BODY     # PR_BODY, COMMIT_MESSAGES, BLANK
  merge_commit_title: PR_TITLE             # PR_TITLE, MERGE_MESSAGE
  merge_commit_message: PR_BODY            # PR_BODY, PR_TITLE, BLANK

  # Security & Access
  allow_update_branch: true
  web_commit_signoff_required: false
  allow_forking: true

  github_pages:
    enabled: false

  # Branch Protection Rules
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

This project uses Nix for reproducible development environments and `just` as a command runner.

```bash
# Enter the dev shell
nix develop
# Or if you use direnv: direnv allow

# Run checks (format, lint, test)
just check

# Build locally
just build
```

## License

MIT
