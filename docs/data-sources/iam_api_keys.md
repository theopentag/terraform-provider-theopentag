---
page_title: "Data Source: theopentag_iam_api_keys"
description: |-
  Lists all platform API keys.
---

# theopentag_iam_api_keys

Lists all platform-level API keys. Full key values are never returned — only the `key_prefix` (first 12 characters) is exposed. Requires `admin` role.

## Example Usage

```hcl
data "theopentag_iam_api_keys" "all" {}

output "sql_keys" {
  value = [
    for k in data.theopentag_iam_api_keys.all.api_keys
    : k.name if contains(k.scopes, "sql")
  ]
}

output "unused_keys" {
  value = [
    for k in data.theopentag_iam_api_keys.all.api_keys
    : k.name if k.last_used_at == null
  ]
}
```

## Schema

### Read-Only

- `api_keys` (List of Object) — List of platform API keys. Each object has:
  - `id` (String) — Numeric API key ID.
  - `name` (String) — Human-readable name.
  - `role` (String) — Role: `admin`, `user`, or `viewer`.
  - `scopes` (List of String) — Module scopes. Empty list means all modules.
  - `key_prefix` (String) — First 12 characters of the key for display purposes.
  - `created_by` (String) — Identity that created this key.
  - `last_used_at` (String) — Timestamp of last usage (ISO UTC). Null if never used.
  - `created_at` (String) — Creation timestamp (ISO UTC).
