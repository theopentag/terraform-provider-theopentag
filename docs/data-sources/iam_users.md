---
page_title: "Data Source: theopentag_iam_users"
description: |-
  Lists all platform users.
---

# theopentag_iam_users

Lists all platform users across all modules. Requires `admin` role.

## Example Usage

```hcl
data "theopentag_iam_users" "all" {}

output "admin_emails" {
  value = [for u in data.theopentag_iam_users.all.users : u.email if u.role == "admin"]
}

output "inactive_users" {
  value = [for u in data.theopentag_iam_users.all.users : u.email if u.last_login == null]
}
```

## Schema

### Read-Only

- `users` (List of Object) — List of platform users. Each object has:
  - `email` (String) — User email address.
  - `name` (String) — Display name.
  - `role` (String) — Platform role: `admin`, `user`, or `viewer`.
  - `last_login` (String) — Timestamp of last login (ISO UTC). Null if never logged in.
  - `created_at` (String) — Creation timestamp (ISO UTC).
