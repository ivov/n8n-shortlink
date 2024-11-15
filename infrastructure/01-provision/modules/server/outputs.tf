output "server_ip" {
  description = "Public IP of the server"
  value       = hcloud_server.main.ipv4_address
}
