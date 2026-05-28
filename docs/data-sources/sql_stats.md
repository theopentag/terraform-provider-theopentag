---
page_title: "Data Source: theopentag_sql_stats"
description: |-
  Aggregate backup statistics across all SQL-managed servers.
---

# theopentag_sql_stats

Reads aggregate backup statistics across all managed servers, including total backup count and backup filesystem disk usage.

## Example Usage

```hcl
data "theopentag_sql_stats" "summary" {}

output "total_backups" {
  value = data.theopentag_sql_stats.summary.total_backups
}

output "barman_disk_free_gb" {
  value = data.theopentag_sql_stats.summary.barman_disk_free / 1073741824
}
```

## Schema

### Read-Only

- `total_backups` (Number) — Total number of backups across all servers.
- `total_disk` (String) — Human-readable total backup disk usage.
- `barman_disk_total` (Number) — Total size of the backup data filesystem in bytes.
- `barman_disk_free` (Number) — Free space on the backup data filesystem in bytes.
