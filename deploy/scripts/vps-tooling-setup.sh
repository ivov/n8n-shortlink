#!/bin/bash

# Set up Caddy, Docker, sqlite3, AWS CLI.

set -euo pipefail

# ==================================
#              caddy
# ==================================

sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
systemctl status caddy
sudo mv deploy/Caddyfile /etc/caddy/Caddyfile
sudo systemctl reload caddy

echo "Caddy installed and running as systemd service, Caddyfile copied over"

# ==================================
#              docker
# ==================================

sudo apt install -y docker.io
sudo usermod --append --groups docker ${USER}

echo "Docker installed and running as systemd service"

# sudo apt install -y docker-compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

echo "docker-compose installed"

# ==================================
#          sqlite3 + DB
# ==================================

sudo apt install -y sqlite3
DOT_DIR=$HOME/.n8n-shortlink
DB_PATH=$DOT_DIR/n8n-shortlink.sqlite
mkdir -p $DOT_DIR
[ ! -f "$DB_PATH" ] && sqlite3 "$DB_PATH" "" && echo "Created empty database at $DB_PATH"

echo "App setup complete"

# ==================================
#             aws cli
# ==================================

echo "Installing AWS CLI. Please have your AWS S3 credentials ready..."

curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip"
sudo apt install -y unzip
unzip awscliv2.zip
sudo ./aws/install
rm awscliv2.zip
aws --version
aws configure # enter AWS S3 credentials
