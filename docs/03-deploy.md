# Deploy

This guide explains how to deploy the app to production.

## Initial deploy

When the docker containers are not running yet, copy the compose files and start the containers.

```sh
cd infrastructure/03-deploy
scp -r * n8n-shortlink-infra:~/.n8n-shortlink/deploy/

ssh n8n-shortlink-infra
docker-compose --file ~/.n8n-shortlink/deploy/docker-compose.monitoring.yml up --detach
docker-compose --file ~/.n8n-shortlink/deploy/docker-compose.yml --profile production up --detach
```

Log in to `https://grafana.domain.com` with `admin/admin` credentials and set a new secure password for the Grafana admin user.

## Deploy on release

When the docker containers are already running, release a new version of the image.

1. Ensure local and remote `master` branches are in sync:

```sh
git fetch origin
git status
# -> Your branch is up to date with 'origin/master'.
```

2. Create a tag (following semver) and push it to remote:

```sh
git tag v1.0.4
git push origin v1.0.4
```

This tag push triggers the [`release` workflow](https://github.com/ivov/n8n-shortlink/actions/workflows/release.yml) to build an ARM64 Docker image. The new image will be [listed](https://github.com/ivov/n8n-shortlink/pkgs/container/n8n-shortlink) on GHCR.

Watchtower polls for new versions of this image every six hours. On discovering a new version of this image, Watchtower will update the container to the new image version.

To prompt Watchtower to check immediately:

```sh
# on a terminal
ssh n8n-shortlink-infra
docker logs -f watchtower

# on another terminal
ssh n8n-shortlink-infra
docker exec watchtower /watchtower --run-once n8n-shortlink
```
