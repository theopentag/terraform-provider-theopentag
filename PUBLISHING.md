# Publishing terraform-provider-theopentag to the Terraform Registry

## Prerequisites

- GitHub account with access to the \*\*theopentag\*\* organization
- Repository named exactly `terraform-provider-theopentag` under that organization
- A GPG key for signing release artifacts
- `goreleaser` installed (`brew install goreleaser`)

---

## Step 1 — Create the GitHub repository

The Terraform Registry requires the repository to follow the naming convention
`terraform-provider-<NAME>` and be public.

```
https://github.com/theopentag/terraform-provider-theopentag
```

Push this directory as the repository root:

```bash
git init
git remote add origin git@github.com:theopentag/terraform-provider-theopentag.git
git add .
git commit -m "initial provider release"
git push -u origin main
```

---

## Step 2 — Generate and export a GPG signing key

The registry requires all release binaries to be signed.

```bash
# Generate a new key (if you don't have one)
gpg --full-generate-key          # choose RSA 4096, no expiry

# Find the key ID
gpg --list-secret-keys --keyid-format=long

# Export the ASCII-armored private key
gpg --armor --export-secret-keys YOUR_KEY_ID
```

Add the exported private key as a GitHub Actions secret named `GPG_PRIVATE_KEY`.
Add the key passphrase (if set) as `GPG_PASSPHRASE`.

---

## Step 3 — Add the GoReleaser configuration

Create `.goreleaser.yml` in the repository root:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
    flags:
      - -trimpath
    ldflags:
      - "-s -w -X main.version={{.Version}}"
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    binary: "{{ .ProjectName }}_v{{ .Version }}"

archives:
  - format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "{{ .ProjectName }}_{{ .Version }}_SHA256SUMS"
  algorithm: sha256

signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"

release:
  draft: false

changelog:
  disable: true
```

---

## Step 4 — Add the GitHub Actions release workflow

Create `.github/workflows/release.yml`:

```yaml
name: release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        id: import_gpg
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.GPG_PASSPHRASE }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

---

## Step 5 — Sign in to the Terraform Registry and connect the repository

1. Go to [registry.terraform.io](https://registry.terraform.io) and sign in with GitHub.
2. Click **Publish → Provider**.
3. Select the `theopentag/terraform-provider-theopentag` repository.
4. Paste the **GPG public key** (get it with `gpg --armor --export YOUR_KEY_ID`).
5. Click **Publish Provider**.

The registry will watch for new tags automatically after this one-time setup.

---

## Step 6 — Tag and release

The registry indexes **semver tags** pushed to GitHub. Every tag triggers GoReleaser
via the workflow above; the registry polls GitHub releases within a few minutes.

```bash
git tag v0.1.0
git push origin v0.1.0
```

After the GitHub Actions workflow completes, the provider will appear at:

```
registry.terraform.io/theopentag/theopentag
```

---

## Step 7 — Use the published provider

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.6"
    }
  }
}

provider "theopentag" {
  host    = "https://sql.example.com"
  api_key = "bmk_..."
}

resource "theopentag_sql_server_config" "primary" {
  name       = "primary-pg17"
  conninfo   = "host=db.example.com port=5432 user=barman dbname=postgres"
  pg_version = 17
}
```

---

## Troubleshooting

### `404 Not Found` for checksums on `terraform init`

```
Error while installing theopentag/theopentag vX.Y.Z: failed to retrieve authentication
checksums for provider: 404 Not Found returned from github.com
```

The Terraform registry fetches `terraform-provider-theopentag_X.Y.Z_SHA256SUMS` and
`terraform-provider-theopentag_X.Y.Z_SHA256SUMS.sig` from the GitHub release. A 404
means the GoReleaser `signs` step did not run — almost always because
`GPG_PRIVATE_KEY` or `GPG_PASSPHRASE` secrets are not set in the repository.

**Fix:**
1. Generate a GPG key and add it to the registry (Step 2 + Step 5 above).
2. Add `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` to **Settings → Secrets and variables → Actions** in the GitHub repository.
3. Re-tag to trigger a clean GoReleaser run:
   ```bash
   git tag vX.Y.Z+1
   git push origin vX.Y.Z+1
   ```
4. Confirm the GitHub release has both `SHA256SUMS` and `SHA256SUMS.sig` files attached.

### `hashicorp/theopentag` provider lookup error

```
Could not retrieve the list of available versions for provider hashicorp/theopentag
```

This means a resource type is referenced without the provider being declared in
`required_providers`, so Terraform defaults to the `hashicorp` namespace.

**Fix:** ensure `required_providers` declares the `theopentag` local name pointing
to `theopentag/theopentag`:

```hcl
terraform {
  required_providers {
    theopentag = {
      source  = "theopentag/theopentag"
      version = ">=0.0.6"
    }
  }
}

provider "theopentag" { ... }

resource "theopentag_sql_server_config" "primary" { ... }
```

---

## Local development install (without publishing)

```bash
make install
```

Then add a `dev_overrides` block to `~/.terraformrc` so Terraform uses the local
binary instead of fetching from the registry:

```hcl
provider_installation {
  dev_overrides {
    "theopentag/theopentag" = "/Users/<you>/.terraform.d/plugins/registry.terraform.io/theopentag/theopentag/0.1.0/<os>_<arch>"
  }
  direct {}
}
```
