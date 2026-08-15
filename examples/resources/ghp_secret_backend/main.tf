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

# Mount the ghp secrets engine and seed it with a service token.
resource "ghp_secret_backend" "ghp" {
  path     = "ghp"
  base_url = "https://ghp.unkin.net"
  ca_cert  = file("${path.module}/ghp-ca.pem")

  admin_token = var.ghp_admin_token # e.g. sourced from vault_kv_secret

  request_timeout_seconds = 30
}

variable "ghp_admin_token" {
  type      = string
  sensitive = true
}
