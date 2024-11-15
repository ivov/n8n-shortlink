variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "n8n-shortlink-terraform"
}

# Hetzner Cloud

variable "hcloud_token" {
  description = "Hetzner Cloud API Token"
  type        = string
  sensitive   = true
}

variable "allowed_ssh_ips" {
  description = "List of IP addresses allowed to connect via SSH, in CIDR notation"
  type        = list(string)
  sensitive   = true

  validation {
    condition     = can([for ip in var.allowed_ssh_ips : regex("^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$", ip)])
    error_message = "Allowed SSH IPs must be in CIDR notation (e.g., [\"1.2.3.4/32\"])"
  }
}

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

variable "tf_automation_aws_access_key_id" {
  description = "AWS access key ID for terraform-automation IAM user"
  type        = string
  sensitive   = true
}

variable "tf_automation_aws_secret_access_key" {
  description = "AWS secret access key for terraform-automation IAM user"
  type        = string
  sensitive   = true
}

