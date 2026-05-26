# terraform-provider-theopentag

Terraform provider for the [Opentag](https://github.com/theopentag) platform — a web dashboard and REST API for managing cloud infrastructure components including PostgreSQL backup management (SQL module) and platform identity & access control (IAM module).

Manage servers, backup policies, schedules, API keys, and platform users as code. Read live status, backup lists, job history, PostgreSQL metadata, and service discovery as data sources.

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.3"
    }
  }
}

provider "theopentag" {
  host    = "https://platform.example.com"
  api_key = "bmk_your_api_key_here"
}
```

---

## Requirements

- Terraform >= 1.0
- Go 1.22+ (to build from source)
- A running Opentag platform instance with an API key (`admin` or `user` role)

---

## Authentication

All requests require a Bearer API key. Generate one in the platform UI under **API Keys**.

| Argument               | Env var              | Description                          |
|------------------------|----------------------|--------------------------------------|
| `host`                 | `PLATFORM_API_HOST`  | Platform API base URL                |
| `api_key` (sensitive)  | `PLATFORM_API_KEY`   | API key (`bmk_...`)                  |
| `insecure_skip_verify` | —                    | Skip TLS verification (dev only)     |

Using environment variables is recommended in CI:

```bash
export PLATFORM_API_HOST="https://platform.example.com"
export PLATFORM_API_KEY="bmk_your_api_key_here"
terraform apply
```

**Role requirements per operation:**

| Role     | What it can do                                                                          |
|----------|-----------------------------------------------------------------------------------------|
| `viewer` | Read status, backups, schedules, jobs, stats, pg metadata, service discovery            |
| `user`   | viewer + trigger backups, create/update server configs, schedules, SQL API keys         |
| `admin`  | user + manage platform users, platform API keys, delete server configs, audit logs      |

---

## Modules

### SQL Module — PostgreSQL backup management

#### Resources

##### `theopentag_sql_server_config`

Registers a PostgreSQL server with the SQL module and configures its backup settings.

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

| Attribute                      | Required | Description                                               |
|-------------------------------|----------|-----------------------------------------------------------|
| `name`                        | yes      | Server name `[a-zA-Z0-9_-]`. Immutable (forces replace)  |
| `conninfo`                    | yes      | libpq connection string (may include `password=`)         |
| `backup_method`               | no       | `postgres` (default) or `rsync`                          |
| `pg_version`                  | no       | PostgreSQL major version 14–18                            |
| `streaming_archiver`          | no       | Enable streaming WAL archiver                             |
| `create_slot`                 | no       | `auto` or `manual`                                        |
| `retention_policy`            | no       | e.g. `RECOVERY WINDOW OF 14 DAYS`                         |
| `backups_enabled`             | no       | Pause/resume backup execution                             |
| `schedule_enabled`            | no       | Enable auto-created daily schedule (create only)          |

Import: `terraform import theopentag_sql_server_config.primary primary-pg17`

---

##### `theopentag_sql_schedule`

Manages a backup schedule for a SQL-managed PostgreSQL server. All times are UTC.

```hcl
resource "theopentag_sql_schedule" "nightly" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Nightly backup"
  schedule_type = "daily"
  schedule_config = { time = "02:00" }
  enabled = true
}

resource "theopentag_sql_schedule" "weekdays" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Mid-week backup"
  schedule_type = "weekly"
  schedule_config = {
    time = "03:00"
    days = [1, 3]   # 0=Sun … 6=Sat
  }
  enabled = true
}

resource "theopentag_sql_schedule" "monthly" {
  server_name   = theopentag_sql_server_config.primary.name
  schedule_type = "monthly"
  schedule_config = { time = "01:00", day = 1 }
  enabled = true
}
```

Import: `terraform import theopentag_sql_schedule.nightly 42`

---

##### `theopentag_sql_api_key`

Manages a SQL-scoped API key. The full `bmk_...` value is returned only on creation and stored as a sensitive value in state — it cannot be retrieved again from the API.

```hcl
resource "theopentag_sql_api_key" "ci" {
  name = "ci-pipeline"
  role = "user"
}

output "ci_key" {
  value     = theopentag_sql_api_key.ci.key
  sensitive = true
}
```

Both `name` and `role` are immutable — any change forces a new key.

Import: `terraform import theopentag_sql_api_key.ci 7`
Note: `key` will be `null` after import (never re-exposed by the API).

---

#### Data Sources

| Data source                        | Description                                            |
|------------------------------------|--------------------------------------------------------|
| `theopentag_sql_server_status`     | Live health and check results for a server             |
| `theopentag_sql_backups`           | All backups for a server, most recent first            |
| `theopentag_sql_jobs`              | Command queue jobs (filter by server, status, limit)   |
| `theopentag_sql_servers`           | All servers with live status summary                   |
| `theopentag_sql_server_configs`    | All server configurations (read-only list)             |
| `theopentag_sql_stats`             | Aggregate backup count and disk usage                  |
| `theopentag_sql_host_stats`        | Latest CPU/RAM/disk snapshot from the backend host     |
| `theopentag_sql_ssh_key`           | SQL's Ed25519 SSH public key for remote restores       |
| `theopentag_sql_pg_databases`      | PostgreSQL databases on a managed server               |
| `theopentag_sql_pg_users`          | PostgreSQL roles on a managed server                   |
| `theopentag_sql_users`             | All SQL API users (admin role required)                |

---

### IAM Module — Platform identity & access control

#### Resources

##### `theopentag_iam_user`

Manages a platform user. Controls access to all modules via role assignment.

```hcl
resource "theopentag_iam_user" "alice" {
  email = "alice@example.com"
  name  = "Alice"
  role  = "user"
}
```

| Attribute    | Required | Description                                         |
|-------------|----------|-----------------------------------------------------|
| `email`     | yes      | User email. Immutable (forces replace)              |
| `name`      | no       | Display name                                        |
| `role`      | yes      | Platform role: `admin`, `user`, or `viewer`         |
| `last_login`| computed | Timestamp of last login (ISO UTC)                   |
| `created_at`| computed | Creation timestamp (ISO UTC)                        |

Import: `terraform import theopentag_iam_user.alice alice@example.com`

---

##### `theopentag_iam_api_key`

Manages a platform-level API key. Scopes restrict the key to specific modules. The full key is returned only on creation.

```hcl
resource "theopentag_iam_api_key" "ci" {
  name   = "ci-pipeline"
  role   = "user"
  scopes = ["sql"]
}

resource "theopentag_iam_api_key" "ops" {
  name   = "ops-automation"
  role   = "admin"
  scopes = ["sql", "iam"]
}

output "ci_key" {
  value     = theopentag_iam_api_key.ci.key
  sensitive = true
}
```

| Attribute     | Required | Description                                                   |
|--------------|----------|---------------------------------------------------------------|
| `name`       | yes      | Human-readable name. Immutable (forces replace)               |
| `role`       | yes      | Role: `admin`, `user`, or `viewer`. Immutable (forces replace)|
| `scopes`     | no       | Module scopes: `["sql"]`, `["sql", "iam"]`, etc. Empty = all modules. Immutable (forces replace) |
| `key`        | computed | Full key value (`bmk_...`). Sensitive, set on creation only   |
| `key_prefix` | computed | First 12 characters for display                               |
| `created_by` | computed | Identity that created this key                                |
| `created_at` | computed | Creation timestamp (ISO UTC)                                  |

Import: `terraform import theopentag_iam_api_key.ci 5`
Note: `key` will be `null` after import.

---

#### Data Sources

##### `theopentag_iam_users`

Lists all platform users (admin role required).

```hcl
data "theopentag_iam_users" "all" {}

output "admins" {
  value = [for u in data.theopentag_iam_users.all.users : u.email if u.role == "admin"]
}
```

##### `theopentag_iam_api_keys`

Lists all platform API keys. Full key values are never returned.

```hcl
data "theopentag_iam_api_keys" "all" {}
```

##### `theopentag_iam_service_discovery`

Reads registered service endpoints from the platform's Service Discovery. Returns live status and API base URLs for all registered modules.

```hcl
data "theopentag_iam_service_discovery" "platform" {}

output "sql_endpoint" {
  value = one([
    for s in data.theopentag_iam_service_discovery.platform.services
    : s.endpoint if s.name == "sql"
  ])
}
```

Each service: `name`, `endpoint`, `status` (`healthy`/`degraded`/`unhealthy`), `version`

---

## Complete Example

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.3"
    }
  }
}

provider "theopentag" {
  host    = "https://platform.example.com"
  api_key = var.platform_api_key
}

# SQL — register a PostgreSQL server and configure backups
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

# IAM — provision a service account user and scoped API key
resource "theopentag_iam_user" "ops" {
  email = "ops-bot@example.com"
  name  = "Ops Bot"
  role  = "user"
}

resource "theopentag_iam_api_key" "ci" {
  name   = "ci-pipeline"
  role   = "user"
  scopes = ["sql"]
}

data "theopentag_sql_server_status" "primary" {
  server_name = theopentag_sql_server_config.primary.name
}

output "server_ok"  { value = data.theopentag_sql_server_status.primary.check_ok }
output "ci_api_key" { value = theopentag_iam_api_key.ci.key; sensitive = true }
```

---

## Development

### Build and run locally

```bash
# Build
make build

# Install to ~/.terraform.d/plugins/
make install
```

Add a `dev_overrides` block to `~/.terraformrc` to use the local binary without `terraform init`:

```hcl
provider_installation {
  dev_overrides {
    "theopentag/theopentag" = "/path/to/opentag-cloud/terraform"
  }
  direct {}
}
```

Then run `terraform plan` / `terraform apply` directly (skip `terraform init`).

### Verify

```bash
go build ./...
go vet ./...
go test ./...
```

### Release

Tag a semver version — GoReleaser builds and signs cross-platform binaries automatically via GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

See [PUBLISHING.md](PUBLISHING.md) for the full registry publish workflow.
