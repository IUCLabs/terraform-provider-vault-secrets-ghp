# A role that mints short-lived agent tokens bound to a ghp App installation,
# scoped to contents and pull requests. Reading ghp/creds/ci returns a
# lease-bound token that ghp revokes when the lease ends.
resource "ghpvaultsecret_secret_role" "ci" {
  backend         = ghpvaultsecret_secret_backend.ghp.path
  name            = "ci"
  token_type      = "agent"
  installation_id = 12345
  repositories    = ["unkin/prodenv"]
  scopes          = ["contents:read", "pull_requests:write"]

  ttl     = 3600  # 1h
  max_ttl = 28800 # 8h
}
