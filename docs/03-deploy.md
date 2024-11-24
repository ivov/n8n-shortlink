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
git tag v1.2.3
git push origin v1.2.3
```

This tag push triggers the [`release` workflow](https://github.com/ivov/n8n-shortlink/actions/workflows/release.yml). On completion, the new image is [listed](https://github.com/ivov/n8n-shortlink/pkgs/container/n8n-shortlink) on GHCR. Watchtower will discover the image and deploy it to production.

Watchtower polls every six hours. To prompt Watchtower to check immediately:

```sh
ssh n8n-shortlink-infra "docker kill --signal=SIGHUP watchtower"
```

Then wait for Watchtower to pull the new image and start the container, and check that the new version is deployed:

```sh
ssh n8n-shortlink-infra "docker ps | grep n8n-shortlink"
```
