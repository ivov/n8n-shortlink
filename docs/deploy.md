# Deployment

## 1. Initial setup

Create SSH key pair and AES encryption key.

```sh
ssh-keygen -t ed25519 -C "<email>" -f ~/.ssh/id_ed25519_shortlink
openssl rand -out ~/.keys/n8n-shortlink-backup-secret.key 32
```

Deploy an Ubuntu ARM64 VPS. Export IP address locally.

```sh
export VPS_IP_ADDRESS=<vps-ip-address>
```

## 2. VPS user setup

Ensure public key is present in VPS, usually added during creation.

```sh
# check if present
ssh -i ~/.ssh/id_ed25519_shortlink root@$VPS_IP_ADDRESS "cat ~/.ssh/authorized_keys"

# else copy over
ssh-copy-id -i ~/.ssh/id_ed25519_shortlink.pub root@$VPS_IP_ADDRESS
```

Run user setup script:

```sh
scp -i ~/.ssh/id_ed25519_shortlink deploy/scripts/vps-user-setup.sh root@$VPS_IP_ADDRESS:/root
ssh -i ~/.ssh/id_ed25519_shortlink root@$VPS_IP_ADDRESS
bash vps-user-setup.sh
```

Configure SSH access:

```sh
export VPS_USER=<username-from-previous-step>

echo "Host shortlink_vps
  HostName $VPS_IP_ADDRESS
  User $VPS_USER
  IdentityFile ~/.ssh/id_ed25519_shortlink" >> ~/.ssh/config

ssh-add ~/.ssh/id_ed25519_shortlink
```

## 3. VPS system setup

Run system setup script:

```sh
scp -r deploy/. shortlink_vps:~/deploy
ssh shortlink_vps
chmod +x deploy/scripts/*.sh
deploy/scripts/vps-system-setup.sh
```

## 4. Third-party services setup

- **Sentry**: Create a project and note down the **DSN**.
- **AWS S3**: Create a bucket and note down the **bucket name**, **region**, **access key** and **secret**. Set a bucket lifecycle rule to delete files prefixed with `n8n-shortlink-backups` older than 10 days.
- **DNS**: Add A records for the `domain.com` and `grafana.domain.com`.
- **GitHub**: Set up a repository. Create a **personal access token** with `read:packages` scope.

## 5. VPS tooling setup

Run tooling setup script and set up backup cron job:

```sh
ssh shortlink_vps 'mkdir -p ~/.keys'
scp ~/.keys/n8n-shortlink-backup-secret.key shortlink_vps:~/.keys
ssh shortlink_vps

export BUCKET_NAME=<bucket-name-from-step-4>

echo "BUCKET_NAME=$BUCKET_NAME" >> deploy/.config
echo "30 23 * * * $HOME/deploy/scripts/backup.sh" | crontab -
deploy/scripts/vps-tooling-setup.sh
```

## 6. Start services

Run server start script:

```sh
export GITHUB_USER=<user-name-from-step-4>
export GITHUB_REPO=<repo-name-from-step-4>
export GITHUB_TOKEN=<github-token-from-step-4>
export SENTRY_DSN=<sentry-dsn-from-step-4>

echo "GITHUB_USER=$GITHUB_USER" >> deploy/.config
echo "GITHUB_REPO=$GITHUB_REPO" >> deploy/.config
mkdir -p ~/.docker
echo '{
  "auths": {
    "ghcr.io": {
      "auth": "'$(echo -n "$GITHUB_USER:$GITHUB_TOKEN" | base64)'"
    }
  }
}' > ~/.docker/config.json
echo "N8N_SHORTLINK_SENTRY_DSN=$SENTRY_DSN" >> deploy/.env.production

docker network create n8n-shortlink-network
deploy/scripts/start-services.sh
```

Log in to `https://grafana.domain.com` and set a password for the Grafana admin user.

## TODOs

- [ ] Introduce shellcheck
