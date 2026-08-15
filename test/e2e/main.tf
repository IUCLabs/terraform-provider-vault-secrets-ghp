terraform {
  required_providers {
    ghpvaultsecret = {
      source = "git.unkin.net/unkin/ghpvaultsecret"
    }
  }
}

provider "ghpvaultsecret" {
  address = "http://127.0.0.1:8200"
  token   = "root"
}

resource "ghpvaultsecret_secret_backend" "ghp" {
  path        = "ghp"
  description = "ghp token engine (e2e)"
  # Reachable from inside the vault container, where the plugin runs.
  base_url        = "http://ghp:3000"
  tls_skip_verify = true

  # Ephemeral test fixture, matched by the mock ghp MOCKGHP_ADMIN_TOKEN.
  admin_token = "ghpsvc_e2e_fixture"
}

resource "ghpvaultsecret_secret_role" "ci" {
  backend         = ghpvaultsecret_secret_backend.ghp.path
  name            = "ci"
  token_type      = "agent"
  installation_id = 12345
  repositories    = ["unkin/prodenv"]
  scopes          = ["contents:read", "pull_requests:write"]
  ttl             = 3600
  max_ttl         = 86400
}
