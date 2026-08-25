# Onboarding to a hcflow Repository

Welcome! This guide gets you up and running in a repository that uses `hcflow`.

## Prerequisites

- `git` installed
- `gh` (GitHub CLI) installed — [cli.github.com](https://cli.github.com)
- `hcflow` installed (see below)

## 1. Install hcflow

```bash
git clone https://github.com/hickstein/hcflow /tmp/hcflow-src
cd /tmp/hcflow-src
make install    # installs to ~/.local/bin/hcflow
```

Verify:
```bash
hcflow --version
```

## 2. Authenticate with GitHub

```bash
gh auth login
```

## 3. Clone and verify

```bash
git clone https://github.com/<owner>/<repo>
cd <repo>
hcflow doctor    # should show all green
hcflow status    # understand the current state
```

## 4. Daily workflow

### Starting work

```bash
hcflow start feat my-feature
# or
hcflow start fix the-bug
```

This creates and switches to `feat/my-feature`.

### Saving progress

```bash
hcflow save "implement the thing"
hcflow save "fix edge case"
```

Messages don't need to follow any special format — they're just for your reference on the branch.

### Opening a PR

```bash
hcflow pr
```

This will:
1. Show a suggested PR title based on your branch name
2. Ask you to confirm or change it (must be a Conventional Commit)
3. Push and create the PR
4. Show the PR URL and initial CI status

### Monitoring

```bash
hcflow status    # text summary
hcflow ui        # visual dashboard
```

## 5. The PR title matters

The PR title is what goes into `main` after squash merge, and what drives versioning.

Examples:
```
feat(telegram): add callback query handling     → minor bump
fix(auth): reconnect expired sessions           → patch bump
feat(api)!: remove deprecated endpoints        → major bump
```

Format: `type(optional-scope): description`

Valid types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`

## 6. What you don't need to do

- ❌ Choose version numbers
- ❌ Write CHANGELOG entries
- ❌ Create git tags
- ❌ Manage GitHub Releases
- ❌ Follow a specific commit format in your branch

Release Please handles all of that after your PR is merged.

## 7. Releases

You don't need to do anything special for releases. After your PR is merged:

1. Release Please automatically updates its Release PR
2. When the project owner is ready: `hcflow release --publish`
3. Release Please creates the tag, CHANGELOG update, and GitHub Release

## Troubleshooting

**`hcflow doctor` shows errors:**
Run the suggested fix commands shown in the doctor output.

**PR title rejected:**
Make sure it follows `type: description` or `type(scope): description` format.

**CI failing:**
Check the Actions tab on GitHub. The CI commands are defined in `.hcflow.yml`.
