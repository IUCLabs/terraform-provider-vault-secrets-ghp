terraform {
  required_providers {
    ghp = {
      source  = "git.unkin.net/unkin/vault-secrets-ghp"
      version = "0.0.1"
    }
  }
}

provider "ghp" {
  # address defaults to $VAULT_ADDR, token to $VAULT_TOKEN
}

variable "ghp_admin_token" {
  type      = string
  sensitive = true
}

resource "ghp_secret_backend" "ghp" {
  path        = "ghp"
  base_url    = "https://ghp.unkin.net"
  admin_token = var.ghp_admin_token
}

# Agent role scoped to a ghp App installation for CI.
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

# Proxy (OAuth-backed) role, no installation binding.
resource "ghp_secret_role" "proxy" {
  backend    = ghp_secret_backend.ghp.path
  name       = "proxy"
  token_type = "proxy"
  scopes     = ["contents:read"]
  ttl        = 1800
  max_ttl    = 14400
}
