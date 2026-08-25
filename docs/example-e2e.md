# End-to-End Example

This walks through the complete lifecycle of a feature from start to release.

## Scenario

You're working on `nbot-tg`, a Telegram bot. You want to add callback query handling.

## Setup (once per repo)

```bash
cd nbot-tg
hcflow init
```

Output:
```
hcflow init

✓ git repository
✓ GitHub remote: hickstein/nbot-tg
✓ default branch: main
✓ gh authenticated

✓ created .hcflow.yml
✓ created .github/workflows/hcflow-ci.yml
✓ created .github/workflows/hcflow-release.yml
✓ created .github/pull_request_template.md
✓ created release-please-config.json
✓ created .release-please-manifest.json

✓ hcflow init complete

  Next steps:
  hcflow doctor        — verify configuration
  hcflow start feat X  — begin work
```

```bash
hcflow doctor
```

Output:
```
Repository
  ✓ git installed             git available
  ✓ git repository            valid git repository

GitHub
  ✓ gh installed              gh available
  ✓ remote origin             github.com/hickstein/nbot-tg
  ✓ authenticated             authenticated
  ✓ repository accessible     accessible

Configuration
  ✓ .hcflow.yml               schema v1
  ✓ schema version            v1 (current)

Workflow
  ✓ CI workflow               configured
  ✓ Release Please workflow   configured

Rules
  ✓ merge strategy            squash merge

GitHub
  ✓ PR template               configured

✓ Result: healthy
```

## Starting a feature

```bash
hcflow start feat telegram-callback
```

Output:
```
✓ main synchronized
✓ created feat/telegram-callback
✓ switched to feat/telegram-callback
```

## Development cycle

```bash
# ... edit code ...

hcflow save "initial callback handler setup"

# ... more edits ...

hcflow save "handle inline keyboard callbacks"

# ... more edits ...

hcflow save "add tests for callback routing"
```

## Status check

```bash
hcflow status
```

Output:
```
nbot-tg

Git
  branch         feat/telegram-callback
  working tree   clean
  ahead/behind   +3 / -0

Pull Request
  no open PR for current branch

Release
  current        v1.4.2
  status         no pending release

hcflow
  schema         v1
  workflows      v1
```

## Open a Pull Request

```bash
hcflow pr
```

Output:
```
Branch
  feat/telegram-callback

Changes
  3 commit(s)

Suggested title: feat(telegram-callback): add callback handler

PR title [press Enter to accept]:
```

_(User presses Enter or types a better title)_

```
Pushing… ✓ pushed

✓ PR #42 created

  https://github.com/hickstein/nbot-tg/pull/42
```

## After CI passes

The PR runs your CI commands:

```yaml
ci:
  commands:
    - make lint
    - make test
```

When CI passes and you're ready, merge the PR via the GitHub UI (squash merge).

The commit on `main` becomes:
```
feat(telegram): add callback query handling
```

## Release

After one or more features are merged, Release Please opens a Release PR:

```bash
hcflow release
```

Output:
```
Release

  current        v1.4.2
  proposed       v1.5.0
  release PR     #47 — chore(main): release 1.5.0
  CI             passed
  status         ready to release

  To publish this release, run: hcflow release --publish
```

When ready:

```bash
hcflow release --publish
```

Output:
```
About to merge Release PR #47 (v1.5.0)
Confirm? [y/N] y
Merging release PR… ✓ done

  Release Please will now:
  • update version in CHANGELOG.md
  • create git tag
  • create GitHub Release
  Run 'hcflow status' to monitor progress.
```

**Result**: GitHub Release `v1.5.0` appears automatically — no manual tagging needed.

## Visual dashboard

At any point:

```bash
hcflow ui
```

Opens `http://127.0.0.1:PORT` with:
- Repository state
- Current PR (with CI status)
- Release state
- Visual pipeline (branch → PR → CI → merge → main → Release Please → Release PR → release)
