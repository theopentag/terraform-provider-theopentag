---
page_title: "Data Source: theopentag_iam_service_discovery"
description: |-
  Reads registered service endpoints from the platform's Service Discovery.
---

# theopentag_iam_service_discovery

Reads registered service endpoints from the platform's Service Discovery (Consul KV). Returns the live status, API base URLs, and deployed versions of all registered platform modules (sql, compute, iam, etc.).

Useful for dynamic configuration — look up a service's base URL rather than hardcoding it, or gate a plan on all services being healthy.

## Example Usage

```hcl
data "theopentag_iam_service_discovery" "platform" {}

output "services" {
  value = data.theopentag_iam_service_discovery.platform.services
}

output "sql_endpoint" {
  value = one([
    for s in data.theopentag_iam_service_discovery.platform.services
    : s.endpoint if s.name == "sql"
  ])
}

output "unhealthy_services" {
  value = [
    for s in data.theopentag_iam_service_discovery.platform.services
    : s.name if s.status != "healthy"
  ]
}
```

## Schema

### Read-Only

- `services` (List of Object) — Registered platform services. Each object has:
  - `name` (String) — Service name (e.g. `sql`, `compute`, `iam`).
  - `endpoint` (String) — API base URL for this service.
  - `status` (String) — Health status: `healthy`, `degraded`, or `unhealthy`.
  - `version` (String) — Deployed version of the service.
- `updated_at` (String) — Timestamp when the service registry was last updated (ISO UTC).
