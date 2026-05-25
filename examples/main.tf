terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.10"
    }
  }
}

provider "theopentag" {
  host    = "https://cloud.example.com"
  api_key = "bmk_your_api_key_here"
  # insecure_skip_verify = false
  # Alternatively, set PLATFORM_API_HOST and PLATFORM_API_KEY environment variables.
}

resource "theopentag_sql_server_config" "primary" {
  name        = "primary-pg17"
  description = "Primary PostgreSQL 17 server"
  conninfo    = "host=db.example.com port=5432 dbname=postgres user=barman password=secret"

  backup_method      = "postgres"
  streaming_conninfo = "host=db.example.com port=5432 dbname=postgres user=streaming_barman password=streamsecret"
  streaming_archiver = true
  create_slot        = "auto"

  sslmode          = "require"
  path_prefix      = "/usr/lib/postgresql/17/bin/"
  pg_version       = 17
  retention_policy = "RECOVERY WINDOW OF 14 DAYS"
  compression      = "bzip2"
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

resource "theopentag_sql_api_key" "monitoring" {
  name = "monitoring"
  role = "viewer"
}

resource "theopentag_sql_api_key" "ci" {
  name = "ci-pipeline"
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

output "server_ok" {
  description = "Whether the primary server is healthy."
  value       = data.theopentag_sql_server_status.primary.ok
}

output "server_check_ok" {
  description = "Whether all critical checks passed."
  value       = data.theopentag_sql_server_status.primary.check_ok
}

output "backup_count" {
  description = "Number of backups available."
  value       = length(data.theopentag_sql_backups.primary.backups)
}

output "monitoring_key_prefix" {
  description = "Display prefix of the monitoring API key."
  value       = theopentag_sql_api_key.monitoring.key_prefix
}

output "monitoring_key" {
  description = "Full monitoring API key (sensitive)."
  value       = theopentag_sql_api_key.monitoring.key
  sensitive   = true
}

output "sql_ssh_public_key" {
  description = "SQL SSH public key for authorizing remote restore operations."
  value       = data.theopentag_sql_ssh_key.sql.public_key
}
