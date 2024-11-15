terraform {
  required_version = ">= 1.0.0"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.45.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.31.0"
    }
  }

  cloud {
    organization = "ivov"
    workspaces {
      name = "n8n-shortlink"
    }
  }
}
