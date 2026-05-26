# CLAUDE.md — terraform-provider-theopentag

Terraform provider for the Opentag platform. Multi-module architecture: **sql** is implemented; **compute** and **iam** are planned. Keep this file as the single source of truth — update it when architecture or naming changes.

---

## Quick commands

```bash
go build ./...                          # compile
go vet ./...                            # static analysis
make build                              # produces ./terraform-provider-theopentag
make install                            # local install into ~/.terraform.d/plugins/
make test                               # go test ./...
make lint                               # golangci-lint
```

---

## Identity

| Key | Value |
|---|---|
| Go module | `github.com/theopentag/terraform-provider-theopentag` |
| Registry address | `registry.terraform.io/theopentag/theopentag` |
| Provider TypeName | `theopentag` |
| GitHub repo | `https://github.com/theopentag/terraform-provider-theopentag` |
| Framework | HashiCorp Plugin Framework v1.13.0, Go 1.22+ |

---

## Provider configuration

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.1"
    }
  }
}

provider "theopentag" {
  host    = "https://platform.example.com"   # or PLATFORM_API_HOST env var
  api_key = "bmk_..."                        # or PLATFORM_API_KEY env var
  # insecure_skip_verify = false
}
```

Environment variables (preferred in CI):
- `PLATFORM_API_HOST` — base URL of the platform API
- `PLATFORM_API_KEY` — API key (`bmk_...` prefix)

---

## Directory structure

```
terraform/
  main.go                          # providerserver.Serve entry point
  go.mod / go.sum
  GNUmakefile
  internal/
    provider/
      provider.go                  # provider Metadata, Schema, Configure, Resources, DataSources
    client/
      client.go                    # shared HTTP client (all modules share one instance)
    modules/
      sql/
        register.go                # exports Resources() and DataSources() for the sql module
        resource_server_config.go
        resource_schedule.go
        resource_api_key.go
        datasource_backups.go
        datasource_host_stats.go
        datasource_jobs.go
        datasource_pg_databases.go
        datasource_pg_users.go
        datasource_server_configs.go
        datasource_server_status.go
        datasource_servers.go
        datasource_ssh_key.go
        datasource_stats.go
        datasource_users.go
  examples/
    main.tf                        # complete usage example
  docs/
    index.md                       # provider-level docs
    resources/                     # per-resource docs
    data-sources/                  # per-data-source docs
```

---

## Naming convention

Terraform splits resource type names at the **first underscore** to find the provider local name.

- Provider TypeName: `"theopentag"` → local name `theopentag`
- Module prefix: `_sql_` for sql module, `_compute_` for compute (future), `_iam_` for iam (future)
- Full pattern: `theopentag_<module>_<type>`

**Never use all-hyphens** (`theopentag-sql-server-config`) — Terraform cannot split on `_` and looks up a non-existent provider.

---

## Resources and data sources

### sql module (`internal/modules/sql/`, package `sql`)

| HCL type | Go file | TypeName suffix |
|---|---|---|
| `resource "theopentag_sql_server_config"` | `resource_server_config.go` | `"_sql_server_config"` |
| `resource "theopentag_sql_schedule"` | `resource_schedule.go` | `"_sql_schedule"` |
| `resource "theopentag_sql_api_key"` | `resource_api_key.go` | `"_sql_api_key"` |
| `data "theopentag_sql_server_status"` | `datasource_server_status.go` | `"_sql_server_status"` |
| `data "theopentag_sql_backups"` | `datasource_backups.go` | `"_sql_backups"` |
| `data "theopentag_sql_jobs"` | `datasource_jobs.go` | `"_sql_jobs"` |
| `data "theopentag_sql_servers"` | `datasource_servers.go` | `"_sql_servers"` |
| `data "theopentag_sql_server_configs"` | `datasource_server_configs.go` | `"_sql_server_configs"` |
| `data "theopentag_sql_stats"` | `datasource_stats.go` | `"_sql_stats"` |
| `data "theopentag_sql_host_stats"` | `datasource_host_stats.go` | `"_sql_host_stats"` |
| `data "theopentag_sql_ssh_key"` | `datasource_ssh_key.go` | `"_sql_ssh_key"` |
| `data "theopentag_sql_pg_databases"` | `datasource_pg_databases.go` | `"_sql_pg_databases"` |
| `data "theopentag_sql_pg_users"` | `datasource_pg_users.go` | `"_sql_pg_users"` |
| `data "theopentag_sql_users"` | `datasource_users.go` | `"_sql_users"` |

All support `terraform import`. Resources: import by name or integer ID. Data sources: read-only.

---

## How provider aggregation works

`provider.go` → `Resources()` / `DataSources()` collect slices from each module's `register.go`:

```go
// provider.go
out = append(out, sqlmod.Resources()...)
// out = append(out, computemod.Resources()...)  // future: compute module
// out = append(out, iammod.Resources()...)       // future: iam module
```

`register.go` (per module) exports two functions only:

```go
func Resources() []func() resource.Resource { ... }
func DataSources() []func() datasource.DataSource { ... }
```

---

## Adding a new module (e.g. compute)

1. Create `internal/modules/compute/` with `package compute`
2. Add `register.go` exporting `Resources()` and `DataSources()`
3. Each resource/datasource uses `req.ProviderTypeName + "_compute_<type>"` as its TypeName
4. In `provider.go`, import as `computemod` and uncomment the two `append` lines
5. Add docs under `docs/resources/` and `docs/data-sources/`

---

## HTTP client

`internal/client/client.go` — shared across all modules. Injected via `Configure`:

```go
c := client.New(host, apiKey, insecure)
resp.DataSourceData = c   // available to all datasources via Configure
resp.ResourceData = c     // available to all resources via Configure
```

Each resource/datasource asserts `req.ProviderData.(*client.Client)` in its own `Configure` method.

---

## Critical constraints

1. **TypeName must contain exactly one `_` before the resource suffix.** Terraform uses the first `_` to find the provider local name. Pattern: `theopentag_<module>_<type>`.

2. **`register.go` is the only file to edit when adding a resource/datasource to a module.** The individual files are self-contained.

3. **`provider.go` is the only file to edit when adding a new module.** It imports each module package and appends its slices.

4. **All imports use** `github.com/theopentag/terraform-provider-theopentag/internal/...` — never the old `terraform-provider-sql` path.

5. **The client is shared.** All modules talk to the same platform API via one `*client.Client`. Do not create per-module clients.

6. **No sensitive values in code or docs.** API keys, passwords, and connection strings in examples must use placeholder values only (e.g. `"bmk_..."`, `"password=secret"`).
