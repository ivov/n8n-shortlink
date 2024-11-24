# n8n-shortlink

Golang app for creating and resolving shortlinks for n8n workflows and URLs.

Little learning project to get familiar with deployment and monitoring best practices.

Live at: https://n8n.to

Features:

- Create + resolve shortlinks for n8n workflows and URLs
- Optionally render n8n workflow shortlinks on canvas
- Vanity URLs and password protection support
- OpenAPI 3.0 spec + Swagger UI playground
- Extensive integration test coverage
- IP-address-based rate limiting

Deployment stack:

- Provisioning with Terraform
- Configuration with Ansible
- Metrics with Prometheus
- Logging with Promtail + Loki
- Monitoring with Grafana
- Caddy as reverse proxy
- Error tracking with Sentry
- Backups with AWS S3 + cronjob
- Releases with GitHub Actions + GHCR
- Deployment with Compose and Watchtower

## Docs

- [`00-develop.md`](docs/00-develop.md)
- [`01-provision.md`](docs/01-provision.md)
- [`02-configure.md`](docs/02-configure.md)
- [`03-deploy.md`](docs/03-deploy.md)
- [`monitor.md`](docs/monitor.md) 
