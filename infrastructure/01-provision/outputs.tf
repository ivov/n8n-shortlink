output "server_ip" {
  description = "Public IP of the Hetzner server"
  value       = module.server.server_ip
}

output "backup" {
  description = "Combined backup configuration and credentials"
  value = {
    credentials = {
      access_key = module.backup.backup_user_access_key
      secret_key = module.backup.backup_user_secret_key
    }
    bucket_name = "${var.project_name}-backup-bucket"
    bucket_region = var.aws_region
  }
  sensitive = true
}
