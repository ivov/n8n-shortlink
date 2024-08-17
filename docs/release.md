# Release

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

## TODOs

- [ ] Shorten build time for `release-on-tag-push.yml`
- [ ] Upgrade dependencies in `release-on-tag-push.yml`
- [ ] Add GitHub release to `release-on-tag-push.yml`
- [ ] Incorporate `check-build-health.yml` into `release-on-tag-push.yml`
