output "backup_user_access_key" {
  description = "AWS access key for backup user"
  value       = aws_iam_access_key.backup.id
}

output "backup_user_secret_key" {
  description = "AWS secret key for backup user"
  value       = aws_iam_access_key.backup.secret
  sensitive   = true
}
