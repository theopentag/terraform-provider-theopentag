---
page_title: "Data Source: theopentag_sql_server_configs"
description: |-
  Lists all servers registered in the SQL management system.
---

# theopentag_sql_server_configs

Lists all servers registered in the SQL management system. Useful for iterating over all managed servers or referencing config values in other resources.

## Example Usage

```hcl
data "theopentag_sql_server_configs" "all" {}

output "pg17_servers" {
  value = [for c in data.theopentag_sql_server_configs.all.server_configs : c.name if c.pg_version == 17]
}
```

## Schema

### Read-Only

- `server_configs` (List of Object) — List of server configurations. Each object has the same attributes as the `theopentag_sql_server_config` resource (excluding `schedule_enabled`):
  - `name` (String)
  - `description` (String)
  - `conninfo` (String)
  - `ssh_command` (String)
  - `backup_method` (String)
  - `archiver` (Boolean)
  - `streaming_conninfo` (String)
  - `streaming_archiver` (Boolean)
  - `create_slot` (String)
  - `slot_name` (String)
  - `path_prefix` (String)
  - `sslmode` (String)
  - `retention_policy` (String)
  - `wal_retention_policy` (String)
  - `minimum_redundancy` (Number)
  - `compression` (String)
  - `backup_compression` (String)
  - `streaming_archiver_batch_size` (Number)
  - `pg_version` (Number)
  - `backups_enabled` (Boolean)
