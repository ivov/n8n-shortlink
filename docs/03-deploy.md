# Deploy

## Old instructions

- **Sentry**: Create a project and note down the **DSN**.

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

