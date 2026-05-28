---
page_title: "Data Source: theopentag_sql_pg_users"
description: |-
  Snapshot of PostgreSQL roles on a managed server.
---

# theopentag_sql_pg_users

Reads a snapshot of PostgreSQL roles on a managed server. The snapshot is refreshed by the SQL backend every 60 seconds via a direct psycopg2 connection.

## Example Usage

```hcl
data "theopentag_sql_pg_users" "primary" {
  server_name = "primary-pg17"
}

output "superusers" {
  value = [
    for u in data.theopentag_sql_pg_users.primary.users : u.username
    if u.is_superuser
  ]
}
```

## Schema

### Required

- `server_name` (String) — Name of the SQL-managed server.

### Read-Only

- `updated_at` (String) — Timestamp of the last snapshot (ISO UTC).
- `users` (List of Object) — List of PostgreSQL roles. Each object has:
  - `username` (String) — Role name.
  - `is_superuser` (Boolean) — Whether the role has superuser privileges.
  - `can_create_roles` (Boolean) — Whether the role can create other roles.
  - `can_create_db` (Boolean) — Whether the role can create databases.
  - `can_login` (Boolean) — Whether the role can log in.
  - `is_replication_user` (Boolean) — Whether the role has replication privilege.
  - `password_valid_until` (String) — Password expiry: `null` = no expiry set, `"infinity"` = never expires, ISO timestamp = expiry date.
  - `member_of_groups` (List of String) — Names of roles this role is a member of.
