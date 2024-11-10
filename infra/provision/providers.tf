provider "hcloud" {
  token = var.hcloud_token
}

provider "aws" {
  region     = var.aws_region
  access_key = var.tf_automation_aws_access_key_id
  secret_key = var.tf_automation_aws_secret_access_key
}
