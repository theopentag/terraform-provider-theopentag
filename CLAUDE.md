# CLAUDE.md — terraform-provider-theopentag

Terraform provider for the Opentag platform. This file is the authoritative reference for AI-assisted development. Update it when architecture, bugs, or patterns change.

---

## Identity

| Key | Value |
|---|---|
| Go module | `github.com/theopentag/terraform-provider-theopentag` |
| Registry | `registry.terraform.io/theopentag/theopentag` |
| Provider TypeName | `theopentag` |
| GitHub repo | `https://github.com/theopentag/terraform-provider-theopentag` |
| Framework | `github.com/hashicorp/terraform-plugin-framework` v1.13.0 |
| Go version | 1.22+ |

---

## Module status

| Module | Package | Status |
|---|---|---|
| sql | `internal/modules/sql` | ✅ implemented |
| iam | `internal/modules/iam` | ✅ implemented |
| compute | `internal/modules/compute` | 🔲 planned |

Both sql and iam are registered in `provider.go`. The compute lines are commented out.

---

## Commands

```bash
make build     # produces ./terraform-provider-theopentag
make install   # copies binary to ~/.terraform.d/plugins/registry.terraform.io/theopentag/theopentag/0.0.1/darwin_arm64/
go build ./... # compile check
go vet ./...   # static analysis
go test ./...  # tests
```

---

## Directory structure

```
main.go                               # var version string + providerserver.Serve
GNUmakefile
.goreleaser.yml                       # GoReleaser v2, formats:["zip"], signs checksum only
.github/workflows/release.yml        # triggers on v* tag push, goreleaser-action@v6 version:"~> v2"
internal/
  provider/provider.go               # ONLY file to edit when adding a new module
  client/client.go                   # ONLY file to edit when adding API methods; shared by all modules
  modules/
    sql/
      register.go                    # ONLY file to edit when adding sql resource/datasource
      resource_server_config.go
      resource_schedule.go
      resource_api_key.go
      datasource_*.go                # 11 datasources
    iam/
      register.go                    # ONLY file to edit when adding iam resource/datasource
      resource_user.go
      resource_api_key.go
      datasource_users.go
      datasource_api_keys.go
      datasource_service_discovery.go
docs/
  index.md                           # provider-level docs (both modules)
  resources/                         # one .md per resource
  data-sources/                      # one .md per datasource
examples/main.tf
```

---

## Naming rules

- Pattern: `theopentag_<module>_<type>` — the first `_` is how Terraform finds the provider local name
- Never use hyphens: `theopentag-sql-...` breaks provider resolution
- TypeName set in `Metadata()`: `resp.TypeName = req.ProviderTypeName + "_sql_server_config"`
- Module prefix: `_sql_`, `_iam_`, `_compute_` (future)

---

## All resources and data sources

### sql module

| HCL name | File | Import key |
|---|---|---|
| `theopentag_sql_server_config` | `resource_server_config.go` | server name (string) |
| `theopentag_sql_schedule` | `resource_schedule.go` | integer ID |
| `theopentag_sql_api_key` | `resource_api_key.go` | integer ID |
| `data.theopentag_sql_server_status` | `datasource_server_status.go` | — |
| `data.theopentag_sql_backups` | `datasource_backups.go` | — |
| `data.theopentag_sql_jobs` | `datasource_jobs.go` | — |
| `data.theopentag_sql_servers` | `datasource_servers.go` | — |
| `data.theopentag_sql_server_configs` | `datasource_server_configs.go` | — |
| `data.theopentag_sql_stats` | `datasource_stats.go` | — |
| `data.theopentag_sql_host_stats` | `datasource_host_stats.go` | — |
| `data.theopentag_sql_ssh_key` | `datasource_ssh_key.go` | — |
| `data.theopentag_sql_pg_databases` | `datasource_pg_databases.go` | — |
| `data.theopentag_sql_pg_users` | `datasource_pg_users.go` | — |
| `data.theopentag_sql_users` | `datasource_users.go` | — |

### iam module

| HCL name | File | Import key |
|---|---|---|
| `theopentag_iam_user` | `resource_user.go` | email (string) |
| `theopentag_iam_api_key` | `resource_api_key.go` | integer ID |
| `data.theopentag_iam_users` | `datasource_users.go` | — |
| `data.theopentag_iam_api_keys` | `datasource_api_keys.go` | — |
| `data.theopentag_iam_service_discovery` | `datasource_service_discovery.go` | — |

---

## HTTP client patterns

`client.go` contains all API structs and methods. Rules:
- One `*client.Client` shared by all modules — never create per-module clients
- Each resource/datasource receives it via `Configure()`: `req.ProviderData.(*client.Client)`
- For PUT/PATCH requests that require a field in the body, always set it explicitly inside the client method, do not rely on the caller

```go
// CORRECT: always enforce name in the body inside the client method
func (c *Client) UpdateServerConfig(ctx context.Context, name string, req ServerConfig) (*ServerConfig, error) {
    req.Name = name  // belt-and-suspenders: ensures name is always in the JSON body
    data, _, err := c.do(ctx, http.MethodPut, "/api/sql/server-configs/"+name, req)
    ...
}
```

- Use `json:"field,omitempty"` only on truly optional fields. Required body fields must NOT have omitempty.

---

## Computed field rules — critical

These rules prevent "Provider produced inconsistent result after apply" errors.

### When to use `UseStateForUnknown()`
Use on fields that are server-assigned on create and **never change** after that (IDs, created_at, key prefixes, slot names).

```go
// CORRECT: ID never changes after create
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
},
```

### When NOT to use `UseStateForUnknown()`
Do NOT use on fields the server recomputes after every update. If you do, Terraform plans the prior state value, but the API returns a new value → inconsistency error.

```go
// WRONG: next_run_at changes when schedule config changes
"next_run_at": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, // BUG
},

// CORRECT: let it be (known after apply) on updates
"next_run_at": schema.StringAttribute{
    Computed: true,
    // no PlanModifiers
},
```

Fields without `UseStateForUnknown()` in this codebase (intentional):
- `theopentag_sql_server_config`: none currently flagged
- `theopentag_sql_schedule`: `next_run_at` — recomputed by server on every schedule change

### Write-only / create-only fields
Fields consumed only on Create (not tracked server-side) need special Update handling — use plan value, not state value:

```go
// resource_server_config.go Update():
// schedule_enabled is only sent on Create; server never returns it.
// Use plan value when known, fall back to state when unknown.
if !plan.ScheduleEnabled.IsUnknown() {
    newState.ScheduleEnabled = plan.ScheduleEnabled
} else {
    newState.ScheduleEnabled = state.ScheduleEnabled
}
```

---

## Adding a new resource (checklist)

1. Create `internal/modules/<module>/resource_<type>.go`
   - `var _ resource.Resource = &<type>Resource{}`
   - `var _ resource.ResourceWithImportState = &<type>Resource{}` (if importable)
   - `Metadata()`: `resp.TypeName = req.ProviderTypeName + "_<module>_<type>"`
   - `Configure()`: assert `req.ProviderData.(*client.Client)`
   - Apply computed field rules above
2. Add constructor to `internal/modules/<module>/register.go`
3. Add API method(s) to `internal/client/client.go`
4. Add `docs/resources/<module>_<type>.md`
5. Update `docs/index.md` table
6. Update `README.md` table
7. Run `go build ./... && go vet ./...`

## Adding a new data source (checklist)

Same as above but:
- `var _ datasource.DataSource = &<type>DataSource{}`
- Only implement `Read()`, no Create/Update/Delete
- Add to `docs/data-sources/<module>_<type>.md`
- All attributes are `Computed: true`

## Adding a new module (checklist)

1. `internal/modules/<module>/` with `package <module>`
2. `register.go` exporting `Resources() []func() resource.Resource` and `DataSources() []func() datasource.DataSource`
3. In `provider.go`: add import alias `<module>mod` and two `append` lines
4. Add API structs/methods to `client.go`
5. Update `docs/index.md`, `README.md` with new module section

---

## API endpoint map

| Client method | HTTP | Path |
|---|---|---|
| `GetServerConfig` | GET | `/api/sql/server-configs/:name` |
| `CreateServerConfig` | POST | `/api/sql/server-configs` |
| `UpdateServerConfig` | PUT | `/api/sql/server-configs/:name` |
| `SetBackupsEnabled` | PATCH | `/api/sql/server-configs/:name/backups-enabled` |
| `DeleteServerConfig` | DELETE | `/api/sql/server-configs/:name` |
| `GetSchedule` | GET | `/api/sql/schedules/:id` |
| `CreateSchedule` | POST | `/api/sql/schedules` |
| `UpdateSchedule` | PUT | `/api/sql/schedules/:id` |
| `DeleteSchedule` | DELETE | `/api/sql/schedules/:id` |
| `ListAPIKeys` | GET | `/api/sql/api-keys` |
| `CreateAPIKey` | POST | `/api/sql/api-keys` |
| `DeleteAPIKey` | DELETE | `/api/sql/api-keys/:id` |
| `CreateIAMUser` | POST | `/api/iam/users` |
| `GetIAMUser` | GET | `/api/iam/users/:email` |
| `UpdateIAMUser` | PATCH | `/api/iam/users/:email` |
| `DeleteIAMUser` | DELETE | `/api/iam/users/:email` |
| `ListIAMUsers` | GET | `/api/iam/users` |
| `CreateIAMAPIKey` | POST | `/api/iam/api-keys` |
| `ListIAMAPIKeys` | GET | `/api/iam/api-keys` |
| `DeleteIAMAPIKey` | DELETE | `/api/iam/api-keys/:id` |
| `GetIAMServiceDiscovery` | GET | `/api/iam/service-discovery` |

---

## Local development

```bash
# Build and point dev_overrides at the project root
make build

# ~/.terraformrc
provider_installation {
  dev_overrides {
    "theopentag/theopentag" = "/path/to/opentag-cloud/terraform"
  }
  direct {}
}

# Terragrunt: skip auto-init (dev_overrides bypass registry, init is not needed)
export TERRAGRUNT_NO_AUTO_INIT=true
terragrunt plan
```

`dev_overrides` bypasses registry lookup during plan/apply. `terraform init` still hits the registry — always skip it with `TERRAGRUNT_NO_AUTO_INIT=true` or `--terragrunt-no-auto-init` when working locally.

---

## Release process

```bash
git tag v0.x.y
git push origin v0.x.y
# GitHub Actions: .github/workflows/release.yml triggers automatically
# GoReleaser builds 6 platform zips + SHA256SUMS + SHA256SUMS.sig
# Terraform Registry indexes within ~30 min
```

Required GitHub secrets: `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`.
The GPG public key must be registered at `registry.terraform.io` → namespace `theopentag` → GPG Keys.
The fingerprint in the release sig must match the registered key ID exactly.

Expected release assets (Terraform Registry requires all of these):
```
terraform-provider-theopentag_X.Y.Z_darwin_amd64.zip
terraform-provider-theopentag_X.Y.Z_darwin_arm64.zip
terraform-provider-theopentag_X.Y.Z_linux_amd64.zip
terraform-provider-theopentag_X.Y.Z_linux_arm64.zip
terraform-provider-theopentag_X.Y.Z_windows_amd64.zip
terraform-provider-theopentag_X.Y.Z_windows_arm64.zip
terraform-provider-theopentag_X.Y.Z_SHA256SUMS
terraform-provider-theopentag_X.Y.Z_SHA256SUMS.sig
```

Binary inside each zip must be named `terraform-provider-theopentag_vX.Y.Z` (with `v` prefix).

---

## Known bugs fixed (do not reintroduce)

| Bug | Root cause | Fix location |
|---|---|---|
| `name` missing from PUT body | `json:"name,omitempty"` + caller not setting Name | `client.go UpdateServerConfig`: always set `req.Name = name` |
| `schedule_enabled` null after update | `UseStateForUnknown()` on write-only field; state was null, plan had `true` | `resource_server_config.go Update()`: use plan value when known |
| `next_run_at` inconsistency after schedule update | `UseStateForUnknown()` on a server-recomputed field | `resource_schedule.go`: no `PlanModifiers` on `next_run_at` |

---

## Invariants — never violate

- TypeName pattern: `theopentag_<module>_<type>` — exactly two underscores minimum
- `register.go` is the only file touched when wiring a new resource/datasource into a module
- `provider.go` is the only file touched when wiring a new module into the provider
- All imports: `github.com/theopentag/terraform-provider-theopentag/internal/...`
- No real secrets in any file — placeholders only (`bmk_...`, `password=secret`)
- `var version string` must exist in `main.go` (ldflags target)
- Do not create per-module HTTP clients — one `*client.Client` for everything
