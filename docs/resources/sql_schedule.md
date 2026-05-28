---
page_title: "Resource: theopentag_sql_schedule"
description: |-
  Manages a backup schedule for a SQL-managed PostgreSQL server.
---

# theopentag_sql_schedule

Manages a backup schedule for a SQL-managed PostgreSQL server. Schedules trigger automated backups at the configured cadence. All times are UTC.

## Example Usage

### Daily backup at 02:30 UTC

```hcl
resource "theopentag_sql_schedule" "nightly" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Nightly backup"
  schedule_type = "daily"
  schedule_config = {
    time = "02:30"
  }
  enabled = true
}
```

### Weekly backup on Monday and Wednesday at 03:00 UTC

```hcl
resource "theopentag_sql_schedule" "weekly" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Weekly backup"
  schedule_type = "weekly"
  schedule_config = {
    time = "03:00"
    days = [1, 3]  # 0=Sun, 1=Mon, ..., 6=Sat
  }
  enabled = true
}
```

### Monthly backup on the 1st at 01:00 UTC

```hcl
resource "theopentag_sql_schedule" "monthly" {
  server_name   = theopentag_sql_server_config.primary.name
  schedule_type = "monthly"
  schedule_config = {
    time = "01:00"
    day  = 1
  }
  enabled = true
}
```

### One-time backup

```hcl
resource "theopentag_sql_schedule" "once" {
  server_name   = theopentag_sql_server_config.primary.name
  schedule_type = "once"
  schedule_config = {
    run_at = "2025-12-01T02:00:00"
  }
  enabled = true
}
```

## Schema

### Required

- `server_name` (String) — Name of the SQL-managed server this schedule applies to.
- `schedule_type` (String) — Schedule cadence: `once`, `daily`, `weekly`, or `monthly`.
- `schedule_config` (Block) — Schedule configuration. Required fields depend on `schedule_type`:
  - `run_at` (String) — ISO datetime for `once` schedules (e.g. `2025-12-01T02:00:00`).
  - `time` (String) — Time in `HH:MM` UTC for `daily`, `weekly`, and `monthly` schedules.
  - `days` (List of Number) — Days of the week for `weekly` schedules. `0`=Sun, `1`=Mon, …, `6`=Sat.
  - `day` (Number) — Day of the month (1–31) for `monthly` schedules.

### Optional

- `label` (String) — Human-readable label.
- `enabled` (Boolean) — Whether the schedule is active. Default: `true`.

### Read-Only

- `id` (String) — Schedule ID.
- `next_run_at` (String) — Next scheduled run time (ISO UTC).
- `last_run_at` (String) — Last run time (ISO UTC).
- `created_at` (String) — Creation timestamp (ISO UTC).

## Import

Import by schedule ID:

```shell
terraform import theopentag_sql_schedule.nightly 42
```
