#!/bin/bash

# Upgrade system, enable firewall, install fail2ban, set up unattended upgrades.

set -euo pipefail

# ==================================
#             admin
# ==================================

sudo timedatectl set-timezone Europe/Berlin

# ==================================
#            upgrades
# ==================================

export DEBIAN_FRONTEND=noninteractive
sudo apt update
sudo apt upgrade -y

# ==================================
#              ufw
# ==================================

sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
echo "y" | sudo ufw enable

# ==================================
#             fail2ban
# ==================================

sudo apt-get install fail2ban -y
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# ==================================
#        unattended upgrades
# ==================================

sudo apt install unattended-upgrades -y
sudo dpkg-reconfigure -f noninteractive unattended-upgrades

echo "System setup complete. Rebooting..."

sudo reboot