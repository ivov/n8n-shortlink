#!/bin/bash

# Create sudo user, grant them SSH access, harden SSH config.

set -euo pipefail

read -p "Enter username: " USERNAME
echo

read -sp "Enter password: " PASSWORD
echo

read -sp "Confirm password: " PASSWORD_CONFIRM
echo

if [ "$PASSWORD" != "$PASSWORD_CONFIRM" ]; then
    echo "Error: Passwords do not match. Please try again."
    exit 1
fi

# ==================================
#           non-root user
# ==================================

sudo adduser --disabled-password --gecos "" "$USERNAME"
echo "${USERNAME}:${PASSWORD}" | sudo chpasswd
sudo usermod --append --groups sudo "$USERNAME"

echo "Created user $USERNAME and added them to sudo group"

# ==================================
#    copy SSH public key to user
# ==================================

mkdir -p /home/$USERNAME/.ssh
chmod 700 /home/$USERNAME/.ssh
cp /root/.ssh/authorized_keys /home/$USERNAME/.ssh/authorized_keys
chown -R $USERNAME:$USERNAME /home/$USERNAME/.ssh
chmod 600 /home/$USERNAME/.ssh/authorized_keys

echo "Copied SSH public key for user $USERNAME"

# ==================================
#           harden auth
# ==================================

sudo sed -i '/^#?PermitRootLogin/c\PermitRootLogin no' /etc/ssh/sshd_config
sudo sed -i '/^#?PasswordAuthentication/c\PasswordAuthentication no' /etc/ssh/sshd_config
sudo systemctl restart ssh
echo "Disabled SSH login for root user"
echo "Disabled password auth for SSH"

echo "User setup complete. Please exit and SSH back in as $USERNAME"