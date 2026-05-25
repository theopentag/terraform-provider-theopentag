---
page_title: "Resource: theopentag_sql_api_key"
description: |-
  Manages a SQL API key.
---

# theopentag_sql_api_key

Manages a SQL API key. The full key value (`bmk_...`) is returned only on creation and stored in Terraform state as a sensitive value. It cannot be retrieved again from the API.

## Example Usage

```hcl
resource "theopentag_sql_api_key" "ci" {
  name = "ci-pipeline"
  role = "user"
}

output "ci_api_key" {
  value     = theopentag_sql_api_key.ci.key
  sensitive = true
}
```

## Schema

### Required

- `name` (String) — Human-readable name. Immutable — changing this forces a new resource.
- `role` (String) — Access role: `admin`, `user`, or `viewer`. Immutable — changing this forces a new resource.

| Role | Permissions |
|---|---|
| `viewer` | Read status, backups, schedules, jobs, stats, pg metadata |
| `user` | viewer + trigger backups, manage server configs and schedules |
| `admin` | user + manage API keys, users, delete server configs, view audit logs |

### Read-Only

- `id` (String) — API key ID.
- `key` (String, Sensitive) — Full API key (`bmk_...`). Available only immediately after creation. Will be `null` after import.
- `key_prefix` (String) — First 12 characters of the key for display purposes.
- `created_by` (String) — Identity that created this key.
- `last_used_at` (String) — Timestamp of last use (ISO UTC).
- `created_at` (String) — Creation timestamp (ISO UTC).

## Import

Import by numeric key ID. Note: the `key` attribute will be `null` after import since the full key is never re-exposed by the API.

```shell
terraform import theopentag_sql_api_key.ci 7
```
