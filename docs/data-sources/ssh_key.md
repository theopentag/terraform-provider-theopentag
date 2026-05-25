---
page_title: "Data Source: theopentag_sql_ssh_key"
description: |-
  Reads the SSH public key used by SQL for remote restore operations.
---

# theopentag_sql_ssh_key

Reads the Ed25519 SSH public key used by SQL for remote restore operations. Add this key to `~/.ssh/authorized_keys` on any PostgreSQL server that SQL needs to restore to.

## Example Usage

```hcl
data "theopentag_sql_ssh_key" "sql" {}

output "sql_public_key" {
  value = data.theopentag_sql_ssh_key.sql.public_key
}
```

### Add the key to a remote server via a provisioner

```hcl
resource "null_resource" "authorize_sql" {
  triggers = {
    key = data.theopentag_sql_ssh_key.sql.public_key
  }

  provisioner "remote-exec" {
    inline = [
      "echo '${data.theopentag_sql_ssh_key.sql.public_key}' >> ~/.ssh/authorized_keys"
    ]
  }
}
```

## Schema

### Read-Only

- `public_key` (String) — Ed25519 public key in OpenSSH format (e.g. `ssh-ed25519 AAAA... sql`). Null if no key has been generated yet.
