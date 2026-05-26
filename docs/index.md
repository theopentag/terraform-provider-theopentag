---
page_title: "Provider: theopentag"
description: |-
  The theopentag provider manages infrastructure components on the Opentag platform — including PostgreSQL backup management (SQL module) and platform identity & access control (IAM module).
---

# theopentag Provider

The **theopentag** provider manages infrastructure components on the [Opentag](https://github.com/theopentag) platform. It covers two modules:

- **SQL** — Register and configure PostgreSQL servers for backup management. Manage backup policies, schedules, and SQL-scoped API keys. Read live server status, backup history, job queues, and PostgreSQL metadata.
- **IAM** — Manage platform users and platform-level API keys. Inspect registered service endpoints via Service Discovery.

## Example Usage

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
```

## Authentication

All API calls require a Bearer token. Generate one in the platform UI under **API Keys** (admin role required), then pass it via the `api_key` argument or the `PLATFORM_API_KEY` environment variable.

**Role requirements:**

| Role     | What it can do                                                                         |
|----------|----------------------------------------------------------------------------------------|
| `viewer` | Read status, backups, schedules, jobs, stats, pg metadata, service discovery           |
| `user`   | viewer + trigger backups, manage server configs, schedules, SQL API keys               |
| `admin`  | user + manage platform users, platform API keys, delete configs, audit logs            |

## Schema

### Optional

- `host` (String) — Base URL of the platform API (e.g. `https://platform.example.com`). Can also be set via `PLATFORM_API_HOST`.
- `api_key` (String, Sensitive) — API key (`bmk_...`). Can also be set via `PLATFORM_API_KEY`.
- `insecure_skip_verify` (Boolean) — Skip TLS certificate verification. Not recommended in production.

## Environment Variables

| Variable             | Description                      |
|----------------------|----------------------------------|
| `PLATFORM_API_HOST`  | Platform API base URL            |
| `PLATFORM_API_KEY`   | API key for authentication       |

## Resources

### SQL Module

| Resource                           | Description                                              |
|------------------------------------|----------------------------------------------------------|
| `theopentag_sql_server_config`     | Registers a PostgreSQL server and configures backups     |
| `theopentag_sql_schedule`          | Manages a backup schedule (daily/weekly/monthly/once)    |
| `theopentag_sql_api_key`           | Manages a SQL-scoped API key                             |

### IAM Module

| Resource                           | Description                                              |
|------------------------------------|----------------------------------------------------------|
| `theopentag_iam_user`              | Manages a platform user (email, name, role)              |
| `theopentag_iam_api_key`           | Manages a platform API key with optional module scopes   |

## Data Sources

### SQL Module

| Data source                        | Description                                              |
|------------------------------------|----------------------------------------------------------|
| `theopentag_sql_server_status`     | Live health and check results for a server               |
| `theopentag_sql_backups`           | All backups for a server                                 |
| `theopentag_sql_jobs`              | Command queue jobs                                       |
| `theopentag_sql_servers`           | All servers with live status summary                     |
| `theopentag_sql_server_configs`    | All server configurations                                |
| `theopentag_sql_stats`             | Aggregate backup count and disk usage                    |
| `theopentag_sql_host_stats`        | CPU/RAM/disk snapshot from the backend host              |
| `theopentag_sql_ssh_key`           | Platform SSH public key for remote restores              |
| `theopentag_sql_pg_databases`      | PostgreSQL databases on a managed server                 |
| `theopentag_sql_pg_users`          | PostgreSQL roles on a managed server                     |
| `theopentag_sql_users`             | All SQL API users                                        |

### IAM Module

| Data source                        | Description                                              |
|------------------------------------|----------------------------------------------------------|
| `theopentag_iam_users`             | All platform users                                       |
| `theopentag_iam_api_keys`          | All platform API keys (without key values)               |
| `theopentag_iam_service_discovery` | Registered service endpoints and their health status     |
