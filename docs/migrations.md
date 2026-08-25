# hcflow Schema Migrations

## Overview

hcflow uses an explicit schema version in `.hcflow.yml`:

```yaml
schema: 1
```

When a new schema version is released, `hcflow upgrade` applies the migration.

## Running upgrades

```bash
hcflow upgrade             # apply all pending migrations
hcflow upgrade --dry-run   # preview changes without applying
```

## Migration guarantees

- **Idempotent**: running twice has the same result as running once
- **Explicit**: every change is described before being applied
- **Non-destructive**: custom workflow content is preserved
- **Versioned**: each migration has a clear from→to schema version

## Current migrations

| From | To | Description |
|------|----|-------------|
| — | — | (no migrations yet — schema v1 is the baseline) |

## Adding a migration (for contributors)

1. Add a new `Migration` to `internal/migrations/migrations.go`:

```go
{
    FromSchema: 1,
    ToSchema:   2,
    Changes: []Change{
        {Kind: "add",    Description: "improved CI permissions"},
        {Kind: "modify", Description: "workflows v1 → v2"},
    },
    Apply: func(dir string, cfg *config.Config, dryRun bool) error {
        // update workflow file
        // update config if needed
        // never silently delete user customizations
        return nil
    },
},
```

2. Write a test in `internal/migrations/migrations_test.go`

3. Update this document

## Schema changelog

### v1 (initial)

- `.hcflow.yml` with project, git, ci, release, github, deploy sections
- `hcflow-ci.yml` workflow
- `hcflow-release.yml` (Release Please)
- `release-please-config.json`
- `.release-please-manifest.json`
- `.github/pull_request_template.md`
