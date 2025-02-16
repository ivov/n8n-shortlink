# Configure

This guide explains how to configure the VPS for the app.

## Setup

1. Install Ansible:

   ```sh
   brew install ansible
   cd infrastructure/02-configure
   ansible all -m ping
   # -> Are you sure you want to continue connecting (yes/no/[fingerprint])? yes
   ```

2. Install Ansible collection for Docker network management:

   ```sh
   ansible-galaxy collection install community.docker
   ansible-galaxy collection list | grep community.docker
   ```

3. In your DNS provider, add A records for the `domain.com` and `grafana.domain.com` pointing to the server's IP address, specified in `02-configure/hosts`.

4. At [Sentry](https://sentry.io/), create a new project `n8n-shortlink` and note down its **DSN**.

## Run

To run all playbooks at once:

```sh
ansible-playbook main.yml
```

During user setup, take note of the non-root user name you create and their `sudo` password. Then add this entry to `~/.ssh/config`, replacing the all-caps values:

```
Host n8n-shortlink-infra
    HostName SERVER_IP_ADDRESS
    User NON_ROOT_USER
    IdentityFile ~/.ssh/id_ed25519_n8n_shortlink_infra
```

Now you can use `ssh n8n-shortlink-infra` or `make vps/login` to SSH in as the non-root user.

## Debug

You can run playbooks one by one for debugging.

### 1. User setup

```sh
ansible-playbook 01-user-setup.yml
```

### 2. System setup

```sh
ansible-playbook 02-system-setup.yml -e "ansible_user=NON_ROOT_USER" --ask-become-pass
```

### 3. Tooling setup

```sh
ansible-playbook 03-tooling-setup.yml -e "ansible_user=NON_ROOT_USER" --ask-become-pass
```

### 4. App dir setup

```sh
ansible-playbook 04-app-dir-setup.yml -e "ansible_user=NON_ROOT_USER"
```

### 5. Backup setup

```sh
ansible-playbook 05-backup-setup.yml -e "ansible_user=NON_ROOT_USER" --ask-become-pass
```
