# terraform-provider-theopentag

Terraform provider for [SQL](https://github.com/theopentag) — a web dashboard and REST API for PostgreSQL backup management.

Manage PostgreSQL servers in SQL, configure backup policies, schedules, and API keys as code. Read live status, backup lists, job history, and PostgreSQL metadata as data sources.

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.6"
    }
  }
}

provider "theopentag" {
  host    = "https://sql.example.com"
  api_key = "bmk_your_api_key_here"
}
```

---

## Requirements

- Terraform >= 1.0
- Go 1.22+ (to build from source)
- A running SQL instance with an API key (`admin` or `user` role)

---

## Authentication

All requests require a Bearer API key. Generate one in the SQL UI under **API Keys**.

**Provider arguments:**

| Argument              | Env var          | Description                          |
|-----------------------|------------------|--------------------------------------|
| `host`                | `PLATFORM_API_HOST`    | SQL API base URL                     |
| `api_key` (sensitive) | `PLATFORM_API_KEY` | API key (`bmk_...`)                  |
| `insecure_skip_verify`| —                | Skip TLS verification (dev only)     |

Using environment variables is recommended in CI:

```bash
export PLATFORM_API_HOST="https://sql.example.com"
export PLATFORM_API_KEY="bmk_your_api_key_here"
terraform apply
```

**Role requirements per operation:**

| Role     | What it can do                                                                |
|----------|-------------------------------------------------------------------------------|
| `viewer` | Read status, backups, schedules, jobs, stats, pg metadata                     |
| `user`   | viewer + trigger backups, create/update server configs and schedules          |
| `admin`  | user + manage API keys, users, delete server configs, view audit logs         |

---

## Resources

### `theopentag_sql_server_config`

Registers a PostgreSQL server with SQL and configures its backup settings.

```hcl
resource "theopentag_sql_server_config" "primary" {
  name               = "primary-pg17"
  description        = "Primary PostgreSQL 17 cluster"
  conninfo           = "host=db.example.com port=5432 user=barman password=secret dbname=postgres"
  streaming_conninfo = "host=db.example.com port=5432 user=streaming_barman password=secret"
  backup_method      = "postgres"
  streaming_archiver = true
  create_slot        = "auto"
  sslmode            = "require"
  path_prefix        = "/usr/lib/postgresql/17/bin/"
  pg_version         = 17
  retention_policy   = "RECOVERY WINDOW OF 14 DAYS"
  minimum_redundancy = 1
  compression        = "bzip2"
  backup_compression = "gzip"
  backups_enabled    = true
  schedule_enabled   = true   # creates a default daily schedule on first apply
}
```

**Key attributes:**

| Attribute                      | Required | Description                                              |
|-------------------------------|----------|----------------------------------------------------------|
| `name`                        | yes      | Server name `[a-zA-Z0-9_-]`. Immutable (forces replace) |
| `conninfo`                    | yes      | libpq connection string (may include `password=`)        |
| `backup_method`               | no       | `postgres` (default) or `rsync`                         |
| `pg_version`                  | no       | PostgreSQL major version 14–18 (default: 17)             |
| `streaming_archiver`          | no       | Enable streaming WAL archiver (default: true)            |
| `create_slot`                 | no       | `auto` or `manual` (default: auto)                       |
| `retention_policy`            | no       | e.g. `RECOVERY WINDOW OF 14 DAYS`                        |
| `backups_enabled`             | no       | Pause/resume backup execution (default: true)            |
| `schedule_enabled`            | no       | Enable the auto-created daily schedule (create only)     |

Import: `terraform import theopentag_sql_server_config.primary primary-pg17`

---

### `theopentag_sql_schedule`

Manages a backup schedule for a SQL-managed PostgreSQL server. All times are UTC.

```hcl
# Daily at 02:30 UTC
resource "theopentag_sql_schedule" "nightly" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Nightly backup"
  schedule_type = "daily"
  schedule_config = {
    time = "02:30"
  }
  enabled = true
}

# Weekly — Mon + Wed at 03:00 UTC
resource "theopentag_sql_schedule" "weekdays" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Mid-week backup"
  schedule_type = "weekly"
  schedule_config = {
    time = "03:00"
    days = [1, 3]  # 0=Sun … 6=Sat
  }
  enabled = true
}

# Monthly on the 1st at 01:00 UTC
resource "theopentag_sql_schedule" "monthly" {
  server_name   = theopentag_sql_server_config.primary.name
  schedule_type = "monthly"
  schedule_config = {
    time = "01:00"
    day  = 1
  }
  enabled = true
}

# One-time
resource "theopentag_sql_schedule" "adhoc" {
  server_name   = theopentag_sql_server_config.primary.name
  schedule_type = "once"
  schedule_config = {
    run_at = "2025-12-01T02:00:00"
  }
  enabled = true
}
```

Import: `terraform import theopentag_sql_schedule.nightly 42`

---

### `theopentag_sql_api_key`

Manages a SQL API key. The full `bmk_...` value is returned only on creation and stored as a sensitive value in state. It cannot be retrieved again from the API.

```hcl
resource "theopentag_sql_api_key" "ci" {
  name = "ci-pipeline"
  role = "user"
}

resource "theopentag_sql_api_key" "monitoring" {
  name = "prometheus"
  role = "viewer"
}

output "ci_key" {
  value     = theopentag_sql_api_key.ci.key
  sensitive = true
}
```

Both `name` and `role` are immutable — any change forces a new key.

Import: `terraform import theopentag_sql_api_key.ci 7`  
Note: `key` will be `null` after import (the full key is never re-exposed by the API).

---

## Data Sources

### `theopentag_sql_server_status`

Live health and check results for a server. Refreshed by the worker every `STATUS_POLL_INTERVAL` seconds (default 30s).

```hcl
data "theopentag_sql_server_status" "primary" {
  server_name = "primary-pg17"
}

output "healthy" {
  value = data.theopentag_sql_server_status.primary.check_ok
}

output "backup_storage" {
  value = data.theopentag_sql_server_status.primary.fields["Backup storage"]
}
```

**Attributes:** `ok`, `check_ok`, `check_items` (list of `{check, status, hint}`), `fields` (map of all SQL status fields), `replication_json`

---

### `theopentag_sql_backups`

All backups for a server, most recent first. Refreshed every 30s by the worker.

```hcl
data "theopentag_sql_backups" "primary" {
  server_name = "primary-pg17"
}

output "done_backups" {
  value = [for b in data.theopentag_sql_backups.primary.backups : b.backup_id if b.status == "DONE"]
}
```

Each backup: `backup_id`, `status` (`DONE`/`FAILED`/`STARTED`/`WAITING_FOR_WALS`/`DONE_WITH_ERRORS`/`EMPTY`), `size`, `begin_time`, `end_time`, `backup_type`, `source` (`manual`/`scheduler`)

---

### `theopentag_sql_jobs`

Command queue jobs. Filter by server, status, and limit.

```hcl
data "theopentag_sql_jobs" "recent_failures" {
  server = "primary-pg17"
  status = "failed"
  limit  = 10
}
```

Each job: `id`, `args_json`, `status`, `exit_code`, `stdout`, `stderr`, `queued_at`, `started_at`, `completed_at`, `schedule_id`, `pg_version`

---

### `theopentag_sql_servers`

All servers with live status summary.

```hcl
data "theopentag_sql_servers" "all" {}

output "unhealthy" {
  value = [for s in data.theopentag_sql_servers.all.servers : s.name if !s.check_ok]
}
```

Each server: `name`, `description`, `active`, `backup_count`, `last_backup`, `disk_usage`, `disk_bytes`, `retention_policy`, `check_ok`, `redundancy_ok`, `redundancy_raw`, `has_active_backup`

---

### `theopentag_sql_server_configs`

All server configurations (read-only list). Useful for iterating without managing configs.

```hcl
data "theopentag_sql_server_configs" "all" {}

output "pg17_servers" {
  value = [for c in data.theopentag_sql_server_configs.all.server_configs : c.name if c.pg_version == 17]
}
```

---

### `theopentag_sql_stats`

Aggregate backup count and backup filesystem disk usage.

```hcl
data "theopentag_sql_stats" "summary" {}

output "total_backups" { value = data.theopentag_sql_stats.summary.total_backups }
output "disk_free_gb"  { value = data.theopentag_sql_stats.summary.barman_disk_free / 1073741824 }
```

**Attributes:** `total_backups`, `total_disk` (human-readable), `barman_disk_total` (backup filesystem total bytes), `barman_disk_free` (backup filesystem free bytes)

---

### `theopentag_sql_host_stats`

Latest CPU/RAM/disk snapshot from the backend host (collected every 15s). Returns an error if no data has been collected yet.

```hcl
data "theopentag_sql_host_stats" "backend" {}
```

**Attributes:** `cpu_percent`, `ram_total`, `ram_used`, `ram_percent`, `disk_total`, `disk_used`, `disk_percent`, `timestamp`

Note: `disk_total`/`disk_used` reflect the backend container's root filesystem, not the backup data disk. Use `theopentag_sql_stats` for backup storage metrics.

---

### `theopentag_sql_ssh_key`

SQL's Ed25519 SSH public key for remote restore operations. Add it to `~/.ssh/authorized_keys` on any server SQL needs to restore to.

```hcl
data "theopentag_sql_ssh_key" "sql" {}

output "sql_public_key" {
  value = data.theopentag_sql_ssh_key.sql.public_key
}
```

---

### `theopentag_sql_pg_databases`

Snapshot of PostgreSQL databases on a managed server. Refreshed every 30s.

```hcl
data "theopentag_sql_pg_databases" "primary" {
  server_name = "primary-pg17"
}

output "db_sizes" {
  value = { for db in data.theopentag_sql_pg_databases.primary.databases : db.database_name => db.size_bytes }
}
```

Each database: `database_name`, `owner`, `encoding`, `collation`, `size_bytes`, `connection_limit`, `is_template`, `allows_connections`

---

### `theopentag_sql_pg_users`

Snapshot of PostgreSQL roles on a managed server. Refreshed every 60s.

```hcl
data "theopentag_sql_pg_users" "primary" {
  server_name = "primary-pg17"
}

output "superusers" {
  value = [for u in data.theopentag_sql_pg_users.primary.users : u.username if u.is_superuser]
}
```

Each user: `username`, `is_superuser`, `can_create_roles`, `can_create_db`, `can_login`, `is_replication_user`, `password_valid_until`, `member_of_groups`

---

### `theopentag_sql_users`

All SQL API users (admin role required).

```hcl
data "theopentag_sql_users" "all" {}

output "admins" {
  value = [for u in data.theopentag_sql_users.all.users : u.email if u.role == "admin"]
}
```

Each user: `email`, `name`, `picture`, `role`, `last_login`, `created_at`

---

## Complete Example

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.6"
    }
  }
}

provider "theopentag" {
  host    = "https://sql.example.com"
  api_key = "bmk_your_api_key_here"
}

resource "theopentag_sql_server_config" "primary" {
  name               = "primary-pg17"
  conninfo           = "host=db.example.com port=5432 user=barman password=secret dbname=postgres"
  streaming_conninfo = "host=db.example.com port=5432 user=streaming_barman password=secret"
  backup_method      = "postgres"
  streaming_archiver = true
  create_slot        = "auto"
  sslmode            = "require"
  path_prefix        = "/usr/lib/postgresql/17/bin/"
  pg_version         = 17
  retention_policy   = "RECOVERY WINDOW OF 14 DAYS"
  minimum_redundancy = 1
  backups_enabled    = true
}

resource "theopentag_sql_schedule" "nightly" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Nightly backup"
  schedule_type = "daily"
  schedule_config = { time = "02:00" }
  enabled = true
}

resource "theopentag_sql_api_key" "ci" {
  name = "ci-pipeline"
  role = "user"
}

data "theopentag_sql_server_status" "primary" {
  server_name = theopentag_sql_server_config.primary.name
}

data "theopentag_sql_ssh_key" "sql" {}

output "server_ok"       { value = data.theopentag_sql_server_status.primary.check_ok }
output "sql_public_key"  { value = data.theopentag_sql_ssh_key.sql.public_key }
output "ci_key"          { value = theopentag_sql_api_key.ci.key; sensitive = true }
```

---

## Development

### Build and install locally

```bash
# Build
make build

# Install to ~/.terraform.d/plugins (version 0.0.1)
make install
```

Add a `dev_overrides` block to `~/.terraformrc` so Terraform uses the local binary:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/theopentag/theopentag" = "/Users/<you>/.terraform.d/plugins/registry.terraform.io/theopentag/theopentag/0.0.1/<os>_<arch>"
  }
  direct {}
}
```

### Build and vet

```bash
go build ./...
go vet ./...
go test ./...
```

### Release

Releases are tagged semver pushes. GoReleaser builds cross-platform binaries and signs checksums. See [PUBLISHING.md](PUBLISHING.md) for the full registry publish workflow.

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions workflow at `.github/workflows/release.yml` runs GoReleaser automatically on tag push.
