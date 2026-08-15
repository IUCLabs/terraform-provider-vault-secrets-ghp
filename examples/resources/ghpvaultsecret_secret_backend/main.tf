terraform {
  required_providers {
    ghpvaultsecret = {
      source = "git.unkin.net/unkin/ghpvaultsecret"
    }
  }
}

provider "ghpvaultsecret" {
  # address / token fall back to VAULT_ADDR / VAULT_TOKEN.
}

# Mount the ghp secrets engine and seed it with a service token.
resource "ghpvaultsecret_secret_backend" "ghp" {
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
