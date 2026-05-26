---
page_title: "Resource: theopentag_iam_user"
description: |-
  Manages a platform user.
---

# theopentag_iam_user

Manages a platform user. Controls access to all modules (SQL, IAM, and future compute) via role assignment. Invite a user by creating this resource — they receive a login email on first access.

## Example Usage

```hcl
resource "theopentag_iam_user" "alice" {
  email = "alice@example.com"
  name  = "Alice"
  role  = "user"
}

resource "theopentag_iam_user" "ops_admin" {
  email = "ops@example.com"
  name  = "Ops Admin"
  role  = "admin"
}
```

## Schema

### Required

- `email` (String) — User email address. Immutable — changing this forces a new resource.
- `role` (String) — Platform role: `admin`, `user`, or `viewer`.

### Optional

- `name` (String) — Display name.

### Read-Only

- `last_login` (String) — Timestamp of last login (ISO UTC). Null if the user has never logged in.
- `created_at` (String) — Creation timestamp (ISO UTC).

## Import

Import by email address:

```shell
terraform import theopentag_iam_user.alice alice@example.com
```
