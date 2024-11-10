output "server_ip" {
  description = "Public IP of the Hetzner server"
  value       = module.server.server_ip
}

output "backup_credentials" {
  description = "AWS credentials for backup user"
  value = {
    access_key = module.backup.backup_user_access_key
    secret_key = module.backup.backup_user_secret_key
  }
  sensitive = true
}
