terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.5"
    }
  }
}

provider "theopentag" {
  host    = "https://cloud.example.com"
  api_key = "bmk_your_api_key_here"
  # insecure_skip_verify = false
  # Alternatively, set PLATFORM_API_HOST and PLATFORM_API_KEY environment variables.
}

# ──────────────────────────────────────────────
# IAM module
# ──────────────────────────────────────────────

resource "theopentag_iam_user" "ops_admin" {
  email = "ops@example.com"
  name  = "Ops Admin"
  role  = "admin"
}

resource "theopentag_iam_user" "developer" {
  email = "dev@example.com"
  name  = "Developer"
  role  = "user"
}

resource "theopentag_iam_api_key" "ci" {
  name   = "ci-pipeline"
  role   = "user"
  scopes = ["sql"]
}

resource "theopentag_iam_api_key" "monitoring" {
  name = "prometheus"
  role = "viewer"
  # scopes omitted = access to all modules
}

data "theopentag_iam_users" "all" {}

data "theopentag_iam_api_keys" "all" {}

data "theopentag_iam_service_discovery" "platform" {}

# ──────────────────────────────────────────────
# SQL module
# ──────────────────────────────────────────────

resource "theopentag_sql_server_config" "primary" {
  name        = "primary-pg17"
  description = "Primary PostgreSQL 17 server"
  conninfo    = "host=db.example.com port=5432 dbname=postgres user=barman password=secret"

  backup_method      = "postgres"
  streaming_conninfo = "host=db.example.com port=5432 dbname=postgres user=streaming_barman password=streamsecret"
  streaming_archiver = true
  create_slot        = "auto"

  sslmode            = "require"
  path_prefix        = "/usr/lib/postgresql/17/bin/"
  pg_version         = 17
  retention_policy   = "RECOVERY WINDOW OF 14 DAYS"
  compression        = "bzip2"
  backup_compression = "gzip"

  minimum_redundancy            = 1
  streaming_archiver_batch_size = 10

  backups_enabled  = true
  schedule_enabled = true
}

resource "theopentag_sql_schedule" "nightly" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Nightly backup"
  schedule_type = "daily"
  schedule_config = {
    time = "02:00"
  }
  enabled = true
}

resource "theopentag_sql_schedule" "weekdays" {
  server_name   = theopentag_sql_server_config.primary.name
  label         = "Weekday backup"
  schedule_type = "weekly"
  schedule_config = {
    time = "03:30"
    days = [1, 2, 3, 4, 5]
  }
  enabled = true
}

resource "theopentag_sql_api_key" "sql_monitoring" {
  name = "sql-monitoring"
  role = "viewer"
}

resource "theopentag_sql_api_key" "sql_ci" {
  name = "sql-ci-pipeline"
  role = "user"
}

data "theopentag_sql_server_status" "primary" {
  server_name = theopentag_sql_server_config.primary.name
}

data "theopentag_sql_backups" "primary" {
  server_name = theopentag_sql_server_config.primary.name
}

data "theopentag_sql_jobs" "recent" {
  server = theopentag_sql_server_config.primary.name
  status = "done"
  limit  = 20
}

data "theopentag_sql_ssh_key" "sql" {}

# ──────────────────────────────────────────────
# Outputs
# ──────────────────────────────────────────────

output "admin_users" {
  description = "Email addresses of all admin users."
  value       = [for u in data.theopentag_iam_users.all.users : u.email if u.role == "admin"]
}

output "sql_service_endpoint" {
  description = "SQL module API endpoint from service discovery."
  value = one([
    for s in data.theopentag_iam_service_discovery.platform.services
    : s.endpoint if s.name == "sql"
  ])
}

output "ci_api_key" {
  description = "CI pipeline API key (sensitive)."
  value       = theopentag_iam_api_key.ci.key
  sensitive   = true
}

output "server_check_ok" {
  description = "Whether all critical checks passed on the primary server."
  value       = data.theopentag_sql_server_status.primary.check_ok
}

output "backup_count" {
  description = "Number of backups available for the primary server."
  value       = length(data.theopentag_sql_backups.primary.backups)
}

output "sql_monitoring_key_prefix" {
  description = "Display prefix of the SQL monitoring API key."
  value       = theopentag_sql_api_key.sql_monitoring.key_prefix
}

output "sql_monitoring_key" {
  description = "Full SQL monitoring API key (sensitive)."
  value       = theopentag_sql_api_key.sql_monitoring.key
  sensitive   = true
}

output "sql_ssh_public_key" {
  description = "SQL SSH public key for authorizing remote restore operations."
  value       = data.theopentag_sql_ssh_key.sql.public_key
}
