# terraform-provider-vault-secrets-ghp

A Terraform/OpenTofu provider that manages the **ghp token secrets engine**
([`vault-plugin-secrets-ghp`](https://git.unkin.net/unkin/vault-plugin-secrets-ghp))
on HashiCorp Vault or OpenBao, so the engine's mount, config, and roles can be
driven declaratively (e.g. from `terraform-vault`).

Source address: `git.unkin.net/unkin/vault-secrets-ghp` (declare it under the
local name `ghp`, so its resources are `ghp_*`).

## Resources

| Resource | Manages |
|----------|---------|
| `ghp_secret_backend` | Mounts the engine at a path and writes its `config` (ghp base URL, TLS, seeded service token). |
| `ghp_secret_role` | A role: `token_type`, `installation_id`, `app_record_id`, `repositories`, `scopes`, `session_prefix`, `ttl`, `max_ttl`. |

## Usage

```hcl
terraform {
  required_providers {
    ghp = {
      source = "git.unkin.net/unkin/vault-secrets-ghp"
    }
  }
}

provider "ghp" {
  # address / token fall back to VAULT_ADDR / VAULT_TOKEN.
}

resource "ghp_secret_backend" "ghp" {
  path        = "ghp"
  base_url    = "https://ghp.unkin.net"
  admin_token = var.ghp_admin_token
}

resource "ghp_secret_role" "ci" {
  backend         = ghp_secret_backend.ghp.path
  name            = "ci"
  token_type      = "agent"
  installation_id = 12345
  repositories    = ["unkin/prodenv"]
  scopes          = ["contents:read", "pull_requests:write"]
  ttl             = 3600
  max_ttl         = 28800
}
```

### Notes

- `admin_token` is write-only: Vault never returns it, so it is preserved in
  Terraform state and does not show drift. `ca_cert` is likewise write-only and
  preserved.
- Writing `config` makes the engine verify the service token against ghp (it must
  authenticate as a ghp admin), so a bad URL or token fails the apply.
- `token_type` defaults to `agent`; agent roles require `installation_id`. Proxy
  roles are OAuth-backed and ignore `installation_id` / `app_record_id`.
- `repositories` and `scopes` are optional; an empty set is open-scoped (all
  repositories / all permissions the installation allows). Scope entries are
  `permission:level` where level is `read` or `write`.

## Import

```sh
terraform import ghp_secret_backend.ghp ghp
terraform import ghp_secret_role.ci ghp/roles/ci
```

## Development

```sh
make build      # build the provider binary
make install    # install into ~/.terraform.d/plugins for local use
make test       # unit tests (race)
make package    # build the release zip
```

Releases are tag-driven (`make patch|minor|major`): a Woodpecker pipeline builds
`terraform-provider-vault-secrets-ghp_<version>_linux_amd64.zip` and PUTs it to
the artifactapi terraform registry
(`.../api/v2/remotes/terraform-unkin/files/unkin/vault-secrets-ghp/<file>`), which
signs it server-side. Install it via the bare `source` address above.
