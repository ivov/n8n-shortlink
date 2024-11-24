variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "hcloud_token" {
  description = "Hetzner Cloud API Token"
  type        = string
  sensitive   = true
}

variable "server_type" {
  description = "Hetzner server type"
  type        = string
}

variable "location" {
  description = "Hetzner datacenter location"
  type        = string
}

variable "ssh_public_key" {
  description = "SSH public key content"
  type        = string
}
