---
page_title: "Provider: theopentag"
description: |-
  The theopentag provider registers and manages PostgreSQL servers in the SQL management system.
---

# theopentag Provider

The **theopentag** provider registers PostgreSQL servers into the [SQL](https://github.com/theopentag) management system and manages their backup configuration, schedules, and API keys as code.

SQL is a web-based platform for PostgreSQL backup management. This provider covers the full server lifecycle: registration, backup policy, schedules, and access control.

## Example Usage

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.6"
    }
  }
}

provider "theopentag" {
  host    = "https://sql.example.com"
  api_key = var.sql_api_key
}
```

## Authentication

All API calls require a Bearer token. Generate one in the SQL UI under **API Keys** (admin role required), then pass it via the `api_key` argument or the `PLATFORM_API_KEY` environment variable.

## Schema

### Optional

- `host` (String) — Base URL of the SQL API (e.g. `https://sql.example.com`). Can also be set via the `PLATFORM_API_HOST` environment variable.
- `api_key` (String, Sensitive) — API key (`bmk_...`). Can also be set via the `PLATFORM_API_KEY` environment variable.
- `insecure_skip_verify` (Boolean) — Skip TLS certificate verification. Not recommended in production.

## Environment Variables

| Variable | Description |
|---|---|
| `PLATFORM_API_HOST` | SQL API base URL |
| `PLATFORM_API_KEY` | API key for authentication |
