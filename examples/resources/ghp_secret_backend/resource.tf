variable "ghp_admin_token" {
  type      = string
  sensitive = true
}

resource "ghp_secret_backend" "ghp" {
  path        = "ghp"
  description = "Short-lived ghp tokens"
  base_url    = "https://ghp.example.com"
  admin_token = var.ghp_admin_token
  ca_cert     = file("${path.module}/ghp-ca.pem")

  request_timeout_seconds = 30
}
