---
page_title: "Data Source: theopentag_sql_servers"
description: |-
  Lists all PostgreSQL servers registered in the SQL management system with their current status.
---

# theopentag_sql_servers

Lists all PostgreSQL servers registered in SQL with live status, backup count, and disk usage.

## Example Usage

```hcl
data "theopentag_sql_servers" "all" {}

output "unhealthy_servers" {
  value = [for s in data.theopentag_sql_servers.all.servers : s.name if !s.check_ok]
}
```

## Schema

### Read-Only

- `servers` (List of Object) — List of servers. Each object has:
  - `name` (String) — Server name.
  - `description` (String) — Human-readable description.
  - `active` (Boolean) — Whether SQL considers the server active.
  - `backup_count` (Number) — Number of available backups.
  - `last_backup` (String) — Timestamp of the most recent backup (ISO UTC).
  - `disk_usage` (String) — Human-readable backup disk usage.
  - `disk_bytes` (Number) — Backup disk usage in bytes.
  - `retention_policy` (String) — Configured retention policy.
  - `check_ok` (Boolean) — Whether all critical checks pass.
  - `redundancy_ok` (Boolean) — Whether minimum redundancy is met.
  - `redundancy_raw` (String) — Raw redundancy string from the backup agent.
  - `has_active_backup` (Boolean) — Whether a backup is currently running.
