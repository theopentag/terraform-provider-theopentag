---
page_title: "Data Source: theopentag_sql_pg_databases"
description: |-
  Snapshot of PostgreSQL databases on a managed server.
---

# theopentag_sql_pg_databases

Reads a snapshot of PostgreSQL databases on a managed server. The snapshot is refreshed by the SQL backend every 30 seconds via a direct psycopg2 connection (not through the backup agent).

## Example Usage

```hcl
data "theopentag_sql_pg_databases" "primary" {
  server_name = "primary-pg17"
}

output "database_sizes" {
  value = {
    for db in data.theopentag_sql_pg_databases.primary.databases :
    db.database_name => db.size_bytes
  }
}
```

## Schema

### Required

- `server_name` (String) — Name of the SQL-managed server.

### Read-Only

- `updated_at` (String) — Timestamp of the last snapshot (ISO UTC).
- `databases` (List of Object) — List of databases. Each object has:
  - `database_name` (String) — Database name.
  - `owner` (String) — Owner role name.
  - `encoding` (String) — Character encoding (e.g. `UTF8`).
  - `collation` (String) — Database collation.
  - `size_bytes` (Number) — Database size in bytes.
  - `connection_limit` (Number) — Maximum connections (`-1` = unlimited).
  - `is_template` (Boolean) — Whether this is a template database.
  - `allows_connections` (Boolean) — Whether connections are allowed.
