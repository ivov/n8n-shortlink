module "server" {
  source          = "./modules/server"
  project_name    = var.project_name
  server_type     = "cax11"
  location        = "nbg1"
  hcloud_token    = var.hcloud_token
  ssh_public_key  = var.ssh_public_key
}

module "backup" {
  source       = "./modules/backup"
  project_name = var.project_name
  bucket_name  = "${var.project_name}-backup-bucket"
}
