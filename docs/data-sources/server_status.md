---
page_title: "Data Source: theopentag_sql_server_status"
description: |-
  Reads the current status of a SQL-managed PostgreSQL server.
---

# theopentag_sql_server_status

Reads the current health status, check results, and raw status fields for a SQL-managed PostgreSQL server. Status is cached by the worker and refreshed every `STATUS_POLL_INTERVAL` seconds (default 30s).

## Example Usage

```hcl
data "theopentag_sql_server_status" "primary" {
  server_name = "primary-pg17"
}

output "primary_healthy" {
  value = data.theopentag_sql_server_status.primary.check_ok
}

output "backup_storage" {
  value = data.theopentag_sql_server_status.primary.fields["Backup storage"]
}
```

## Schema

### Required

- `server_name` (String) — Name of the SQL-managed server.

### Read-Only

- `ok` (Boolean) — Overall server health reported by SQL.
- `check_ok` (Boolean) — Whether all critical checks pass (non-critical checks such as minimum redundancy and WAL age are excluded).
- `check_items` (List of Object) — Individual check results:
  - `check` (String) — Check name.
  - `status` (String) — `OK` or `FAILED`.
  - `hint` (String) — Additional detail or hint.
- `fields` (Map of String) — All SQL status fields as a string map (e.g. `PostgreSQL version`, `Backup storage`, `Current LSN`).
- `replication_json` (String) — Replication status as a raw JSON string (`{current_lsn, client_count, clients: [...]}`). Null if no replication data is available.
