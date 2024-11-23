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


## Old: Release

Ensure local and remote `master` branches are in sync:

```sh
git fetch origin
git status
# -> Your branch is up to date with 'origin/master'.
```

Create tag (following semver) and push it:

```sh
git tag v1.0.0
git push origin v1.0.0
```

Monitor the release on GitHub:

- https://github.com/ivov/n8n-shortlink/actions/workflows/release-on-tag-push.yml

On completion, this release is listed as `latest` on GHCR:

- https://github.com/ivov/n8n-shortlink/pkgs/container/n8n-shortlink

Deploy the release to production:

```sh
ssh shortlink_vps

COMPOSE_PROJECT_NAME=n8n_shortlink docker-compose --file deploy/docker-compose.monitoring.yml down
COMPOSE_PROJECT_NAME=n8n_shortlink docker-compose --file deploy/docker-compose.yml down
deploy/scripts/start-services.sh
```
