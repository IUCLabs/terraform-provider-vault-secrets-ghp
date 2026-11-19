# Terraform Provider for the Vault ghp secrets engine

Manages the [ghp secrets engine](https://github.com/IUCLabs/vault-plugin-secrets-ghp)
on HashiCorp Vault or OpenBao from Terraform or OpenTofu: mount the engine, write
its connection config and service token, and define the roles that mint
short-lived [ghp](https://github.com/goodtune/ghp) tokens.

The provider's resources are prefixed `ghp_`, so declare it under the local name
`ghp`.

## Quick Start

Prebuilt provider archives are on the [releases page](https://github.com/IUCLabs/terraform-provider-vault-secrets-ghp/releases).

```hcl
terraform {
  required_providers {
    ghp = {
      source = "IUCLabs/vault-secrets-ghp"
    }
  }
}

# address and token fall back to VAULT_ADDR and VAULT_TOKEN.
provider "ghp" {}

resource "ghp_secret_backend" "ghp" {
  path        = "ghp"
  base_url    = "https://ghp.example.com"
  admin_token = var.ghp_admin_token
}

resource "ghp_secret_role" "ci" {
  backend         = ghp_secret_backend.ghp.path
  name            = "ci"
  installation_id = 12345
  repositories    = ["example/app"]
  scopes          = ["contents:read", "pull_requests:write"]
  ttl             = 3600
  max_ttl         = 28800
}
```

Applications then read `ghp/creds/ci` from Vault to get a token.

## Requirements

- Terraform 1.x or OpenTofu
- HashiCorp Vault or OpenBao with
  [vault-plugin-secrets-ghp](https://github.com/IUCLabs/vault-plugin-secrets-ghp)
  registered in the plugin catalog
- Go 1.25+ to build the provider

## Usage

See the [docs](docs/index.md):

- [`ghp_secret_backend`](docs/resources/secret_backend.md)
- [`ghp_secret_role`](docs/resources/secret_role.md)

## Development

```sh
make build    # build the provider binary
make install  # install into ~/.terraform.d/plugins for local testing
make test     # unit tests
make lint     # go vet
make docs     # regenerate docs/ from the schema and examples/
```

## License

[MIT](LICENSE)
