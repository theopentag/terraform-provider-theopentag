---
page_title: "Data Source: theopentag_sql_users"
description: |-
  Lists all users with access to the SQL API.
---

# theopentag_sql_users

Lists all users registered in the SQL API. Includes both Google SSO users and the local admin account. Requires admin role.

## Example Usage

```hcl
data "theopentag_sql_users" "all" {}

output "admin_users" {
  value = [for u in data.theopentag_sql_users.all.users : u.email if u.role == "admin"]
}
```

## Schema

### Read-Only

- `users` (List of Object) — List of users. Each object has:
  - `email` (String) — User email (primary identifier).
  - `name` (String) — Display name.
  - `picture` (String) — Avatar URL (Google SSO users only).
  - `role` (String) — Access role: `admin`, `user`, or `viewer`.
  - `last_login` (String) — Timestamp of last login (ISO UTC).
  - `created_at` (String) — Account creation timestamp (ISO UTC).
