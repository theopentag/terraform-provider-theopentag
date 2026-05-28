---
page_title: "Data Source: theopentag_sql_backups"
description: |-
  Lists all backups for a SQL-managed server.
---

# theopentag_sql_backups

Lists all backups for a SQL-managed server. Backup data is refreshed by the worker every 30 seconds.

## Example Usage

```hcl
data "theopentag_sql_backups" "primary" {
  server_name = "primary-pg17"
}

output "latest_backup_id" {
  value = data.theopentag_sql_backups.primary.backups[0].backup_id
}

output "done_backups" {
  value = [for b in data.theopentag_sql_backups.primary.backups : b.backup_id if b.status == "DONE"]
}
```

## Schema

### Required

- `server_name` (String) — Name of the SQL-managed server.

### Read-Only

- `backups` (List of Object) — List of backups, most recent first. Each object has:
  - `backup_id` (String) — Backup identifier (e.g. `20240101T120000`).
  - `status` (String) — `DONE`, `FAILED`, `STARTED`, `WAITING_FOR_WALS`, `DONE_WITH_ERRORS`, or `EMPTY`.
  - `size` (String) — Human-readable backup size.
  - `begin_time` (String) — Backup start time (ISO UTC).
  - `end_time` (String) — Backup end time (ISO UTC).
  - `backup_type` (String) — Backup type (e.g. `full`, `incremental`).
  - `source` (String) — What triggered the backup: `manual` or `scheduler`.
