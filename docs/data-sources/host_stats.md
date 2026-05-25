---
page_title: "Data Source: theopentag_sql_host_stats"
description: |-
  Latest backend host CPU, RAM, and disk snapshot.
---

# theopentag_sql_host_stats

Reads the most recent host resource snapshot from the SQL backend (collected every 15 seconds). Returns an error if no data has been collected yet.

Note: `disk_total` and `disk_used` reflect the **backend container's root filesystem**, not the backup data disk. Use `theopentag_sql_stats.barman_disk_total` / `barman_disk_free` for backup storage metrics.

## Example Usage

```hcl
data "theopentag_sql_host_stats" "backend" {}

output "cpu_percent" {
  value = data.theopentag_sql_host_stats.backend.cpu_percent
}

output "ram_used_gb" {
  value = data.theopentag_sql_host_stats.backend.ram_used / 1073741824
}
```

## Schema

### Read-Only

- `cpu_percent` (Number) — CPU usage percent averaged over the last 15-second window.
- `ram_total` (Number) — Total RAM in bytes.
- `ram_used` (Number) — Used RAM in bytes.
- `ram_percent` (Number) — RAM usage percent.
- `disk_total` (Number) — Total disk size in bytes (backend container root filesystem).
- `disk_used` (Number) — Used disk in bytes.
- `disk_percent` (Number) — Disk usage percent.
- `timestamp` (String) — Snapshot timestamp (ISO UTC).
