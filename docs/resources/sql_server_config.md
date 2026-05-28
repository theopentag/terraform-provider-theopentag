---
page_title: "Resource: theopentag_sql_server_config"
description: |-
  Registers a PostgreSQL server with the SQL management system.
---

# theopentag_sql_server_config

Registers a PostgreSQL server with the SQL management system and configures its backup settings. Once registered, SQL begins monitoring the server, managing WAL archiving, and running backup schedules.

## Example Usage

```hcl
resource "theopentag_sql_server_config" "primary" {
  name              = "primary-pg17"
  description       = "Primary PostgreSQL 17 cluster"
  conninfo          = "host=db.example.com port=5432 user=barman password=secret dbname=postgres"
  streaming_conninfo = "host=db.example.com port=5432 user=streaming_barman password=secret"
  backup_method     = "postgres"
  streaming_archiver = true
  create_slot       = "auto"
  sslmode           = "require"
  retention_policy  = "RECOVERY WINDOW OF 14 DAYS"
  minimum_redundancy = 1
  pg_version        = 17
  backups_enabled   = true
  schedule_enabled  = true
}
```

## Schema

### Required

- `name` (String) — Server name. Allowed characters: `[a-zA-Z0-9_-]`. Immutable — changing this forces a new resource.
- `conninfo` (String) — libpq connection string for the PostgreSQL server. May include `password=`.

### Optional

- `description` (String) — Human-readable description.
- `ssh_command` (String) — SSH command used for rsync backup method (e.g. `ssh postgres@db.example.com`).
- `backup_method` (String) — Backup method: `postgres` (streaming) or `rsync`. Default: `postgres`.
- `archiver` (Boolean) — Enable WAL file archiver. Default: `false`.
- `streaming_conninfo` (String) — Streaming replication connection string.
- `streaming_archiver` (Boolean) — Enable streaming WAL archiver. Default: `true`.
- `create_slot` (String) — Replication slot creation: `auto` or `manual`. Default: `auto`.
- `slot_name` (String) — Replication slot name. Defaults to the server name with hyphens replaced by underscores.
- `path_prefix` (String) — Path to PostgreSQL binaries (e.g. `/usr/lib/postgresql/17/bin/`).
- `sslmode` (String) — SSL mode appended to `conninfo` and `streaming_conninfo`.
- `retention_policy` (String) — Backup retention policy (e.g. `RECOVERY WINDOW OF 14 DAYS`).
- `wal_retention_policy` (String) — WAL retention policy. Always `main`.
- `minimum_redundancy` (Number) — Minimum number of backups to keep. Default: `1`.
- `compression` (String) — WAL compression algorithm (e.g. `bzip2`, `gzip`).
- `backup_compression` (String) — Backup compression algorithm (e.g. `gzip`, `snappy`).
- `streaming_archiver_batch_size` (Number) — WAL files retrieved per streaming archiver batch. Default: `10`.
- `pg_version` (Number) — PostgreSQL major version (14–18). Determines which worker processes this server. Default: `17`.
- `backups_enabled` (Boolean) — Enable or disable backup execution for this server. Default: `true`.
- `schedule_enabled` (Boolean) — Whether the auto-created daily schedule starts enabled. Consumed only on create; not tracked after. Default: `true`.

## Import

Import by server name:

```shell
terraform import theopentag_sql_server_config.primary primary-pg17
```
