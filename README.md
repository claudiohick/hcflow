# hcflow

> Opinionated Git/GitHub workflow orchestrator — small, automated, language-agnostic.

`hcflow` standardises the daily development cycle across multiple repositories by orchestrating Git, GitHub CLI, GitHub Actions, Conventional Commits and Release Please into a low-friction experience.

---

## Philosophy

```
hcflow is a thin orchestration layer — not a new platform.
```

It wraps:

| Tool | Role |
|------|------|
| **Git** | Source of truth for code and history |
| **GitHub CLI (`gh`)** | PRs, releases, repo management |
| **GitHub Actions** | CI and release automation |
| **Conventional Commits** | Structured commit titles on `main` |
| **Release Please** | Versioning, CHANGELOG, tags, GitHub Releases |

hcflow does **not** re-implement any of the above. It provides a consistent UX and operational model on top of them.

---

## Quick Start

### Install

```bash
# Build from source
git clone https://github.com/claudiohick/hcflow
cd hcflow
make install   # installs to ~/.local/bin/hcflow
```

### Bootstrap a repository

```bash
cd your-project
hcflow init    # creates .hcflow.yml, workflows, Release Please config
hcflow doctor  # verify everything looks good
```

### Daily workflow

```bash
# Start a unit of work
hcflow start feat telegram-callback

# ... develop ...

hcflow save "initial implementation"
hcflow save "fix callback tests"

# Open a Pull Request
hcflow status
hcflow pr
```

After CI passes and the PR is merged:

```bash
# Monitor release status
hcflow release

# When ready
hcflow release --publish
```

### Visual dashboard

```bash
hcflow ui    # opens http://127.0.0.1:<port> in your browser
```

---

## Commands

| Command | Purpose |
|---------|---------|
| `hcflow init` | Bootstrap a repository |
| `hcflow start <type> <desc>` | Create a feature branch |
| `hcflow save [msg]` | Commit all current changes |
| `hcflow status` | Show operational state |
| `hcflow pr` | Create a Pull Request |
| `hcflow release` | Inspect or publish a release |
| `hcflow doctor` | Diagnose the configuration |
| `hcflow upgrade` | Upgrade to the latest schema |
| `hcflow ui` | Open the web dashboard |

### Branch types

```
feat    fix    perf    refactor
docs    test   build   ci    chore
```

---

## Configuration: `.hcflow.yml`

```yaml
schema: 1

project:
  name: my-project

git:
  default_branch: main
  merge_strategy: squash

ci:
  enabled: true
  commands:
    - make lint
    - make test

release:
  enabled: true
  provider: release-please
  strategy: semver

github:
  pull_requests: true
  linear_history: true
  required_approvals: 0

deploy:
  enabled: false
```

### Language-agnostic CI

`ci.commands` accepts any shell commands:

```yaml
# Go
ci:
  commands:
    - make lint
    - make test

# Node.js
ci:
  commands:
    - npm ci
    - npm test

# Rust
ci:
  commands:
    - cargo test

# Python
ci:
  commands:
    - make lint
    - pytest
```

---

## Flow

```
hcflow start feat X
      │
      ▼
 local dev
 (any commits)
      │
      ▼
hcflow save "..."
      │
      ▼
hcflow pr
      │ (suggests conventional commit title)
      ▼
 GitHub PR
      │
      ▼
 CI (your commands)
      │
      ▼
 squash merge
      │
 feat(x): description
      │
      ▼
     main
      │
      ▼
Release Please
      │
      ▼
 Release PR
      │
      ▼
hcflow release --publish
      │
      ▼
version • CHANGELOG • tag • GitHub Release
```

### The PR title is the unit of meaning

Local commits during development can be anything:
```
primeira tentativa
fix tests  
ajusta callback
agora funciona
```

The **PR title** is what ends up on `main` after squash merge, and what Release Please uses to compute the next version. It must follow Conventional Commits:

```
feat(telegram): add callback query handling
fix(auth): reconnect expired session
feat(api)!: remove legacy endpoint  ← breaking change → major bump
```

---

## Release Please Integration

hcflow does **not** manage versions itself. Release Please is the authority.

| Commit type | Version bump |
|-------------|-------------|
| `fix:` | patch |
| `feat:` | minor |
| `feat!:` or `BREAKING CHANGE` | major |

After merging feature PRs to `main`, Release Please automatically:
1. Maintains a **Release PR** with the proposed version and CHANGELOG
2. On `hcflow release --publish` (merging the Release PR):
   - Updates version
   - Updates CHANGELOG
   - Creates a git tag
   - Creates a GitHub Release

---

## Architecture

See [docs/architecture.md](docs/architecture.md).

---

## Adding a new collaborator

```bash
git clone https://github.com/owner/repo
cd repo

# Install hcflow (see Install above)

hcflow doctor    # verify setup
hcflow status    # understand current state
hcflow start feat my-first-feature
```

See [docs/onboarding.md](docs/onboarding.md) for detailed onboarding.

---

## License

MIT
