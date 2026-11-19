# Agent tokens bound to a GitHub App installation, limited to one repository.
resource "ghp_secret_role" "ci" {
  backend         = ghp_secret_backend.ghp.path
  name            = "ci"
  token_type      = "agent"
  installation_id = 12345
  repositories    = ["example/app"]
  scopes          = ["contents:read", "pull_requests:write"]

  ttl     = 3600
  max_ttl = 28800
}

# OAuth-backed proxy tokens.
resource "ghp_secret_role" "proxy" {
  backend    = ghp_secret_backend.ghp.path
  name       = "proxy"
  token_type = "proxy"
  scopes     = ["contents:read"]
  ttl        = 1800
  max_ttl    = 14400
}
