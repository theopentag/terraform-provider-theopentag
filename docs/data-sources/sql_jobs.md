---
page_title: "Data Source: theopentag_sql_jobs"
description: |-
  Lists SQL command queue jobs.
---

# theopentag_sql_jobs

Lists SQL command queue jobs. Useful for checking the status of recent backup or delete operations.

## Example Usage

```hcl
data "theopentag_sql_jobs" "recent_failures" {
  server = "primary-pg17"
  status = "failed"
  limit  = 10
}

output "failed_job_ids" {
  value = [for j in data.theopentag_sql_jobs.recent_failures.jobs : j.id]
}
```

## Schema

### Optional

- `server` (String) — Filter jobs by server name.
- `status` (String) — Filter by status: `pending`, `running`, `done`, or `failed`.
- `limit` (Number) — Maximum number of jobs to return (1–200). Default: `50`.

### Read-Only

- `jobs` (List of Object) — List of jobs. Each object has:
  - `id` (String) — Job ID.
  - `cache_key` (String) — Deduplication cache key.
  - `args_json` (String) — JSON array of command arguments (e.g. `["backup","primary-pg17"]`).
  - `status` (String) — `pending`, `running`, `done`, or `failed`.
  - `exit_code` (String) — Process exit code (as string; null if not yet completed).
  - `stdout` (String) — Standard output.
  - `stderr` (String) — Standard error.
  - `queued_at` (String) — Time the job was queued (ISO UTC).
  - `started_at` (String) — Time the job started (ISO UTC).
  - `completed_at` (String) — Time the job completed (ISO UTC).
  - `schedule_id` (String) — ID of the schedule that triggered this job, if any.
  - `pg_version` (String) — PostgreSQL major version this job targets, if any.
