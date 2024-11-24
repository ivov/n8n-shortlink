variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "n8n-shortlink-infra"
}

# Hetzner Cloud

variable "hcloud_token" {
  description = "Hetzner Cloud API Token"
  type        = string
  sensitive   = true
}

# local

variable "ssh_public_key" {
  description = "SSH public key content"
  type        = string
  sensitive   = true
}

# AWS

variable "aws_region" {
  description = "AWS Region for backup resources"
  type        = string
  default     = "eu-central-1"
}

variable "aws_access_key_id" {
  description = "AWS access key ID for backup user"
  type        = string
  sensitive   = true
}

variable "aws_secret_access_key" {
  description = "AWS secret access key for backup user"
  type        = string
  sensitive   = true
}

