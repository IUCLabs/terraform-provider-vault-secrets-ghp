# Import by mount path. admin_token and ca_cert cannot be read back from Vault,
# so set them in configuration and run apply after importing.
terraform import ghp_secret_backend.ghp ghp
