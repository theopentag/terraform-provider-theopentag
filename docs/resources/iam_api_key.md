---
page_title: "Resource: theopentag_iam_api_key"
description: |-
  Manages a platform-level API key with optional module scopes.
---

# theopentag_iam_api_key

Manages a platform-level API key. Scopes restrict the key to specific modules — useful for service accounts that should only access one part of the platform. The full key value (`bmk_...`) is returned only on creation and stored as a sensitive value in state; it cannot be retrieved again from the API.

## Example Usage

```hcl
# SQL-only CI key
resource "theopentag_iam_api_key" "ci" {
  name   = "ci-pipeline"
  role   = "user"
  scopes = ["sql"]
}

# Ops key with access to SQL and IAM
resource "theopentag_iam_api_key" "ops" {
  name   = "ops-automation"
  role   = "admin"
  scopes = ["sql", "iam"]
}

# Unrestricted monitoring key
resource "theopentag_iam_api_key" "monitoring" {
  name = "prometheus"
  role = "viewer"
  # scopes omitted = access to all modules
}

output "ci_key" {
  value     = theopentag_iam_api_key.ci.key
  sensitive = true
}
```

## Schema

### Required

- `name` (String) — Human-readable name. Immutable — changing this forces a new resource.
- `role` (String) — Role: `admin`, `user`, or `viewer`. Immutable — changing this forces a new resource.

### Optional

- `scopes` (List of String) — Module scopes this key has access to (e.g. `["sql"]`, `["sql", "iam"]`). An empty list means the key has access to all modules. Immutable — changing this forces a new resource.

### Read-Only

- `id` (String) — Numeric API key ID.
- `key` (String, Sensitive) — Full key value (`bmk_...`). Set only on creation; preserved in state thereafter.
- `key_prefix` (String) — First 12 characters of the key for display purposes.
- `created_by` (String) — Identity that created this key.
- `last_used_at` (String) — Timestamp of last key usage (ISO UTC). Null if never used.
- `created_at` (String) — Creation timestamp (ISO UTC).

## Import

Import by numeric ID:

```shell
terraform import theopentag_iam_api_key.ci 5
```

Note: `key` will be `null` after import — the full key is never re-exposed by the API.
