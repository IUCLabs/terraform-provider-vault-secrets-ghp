terraform {
  required_providers {
    ghp = {
      source = "IUCLabs/vault-secrets-ghp"
    }
  }
}

# address and token fall back to VAULT_ADDR and VAULT_TOKEN.
provider "ghp" {
  address = "https://vault.example.com:8200"
}

variable "ghp_admin_token" {
  type      = string
  sensitive = true
}

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
