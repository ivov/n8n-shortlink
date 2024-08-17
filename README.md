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

- Metrics with expvar, Prometheus, node exporter, cAdvisor
- Logging with zap, Promtail, Loki
- Monitoring with Grafana
- Caddy as reverse proxy
- Error tracking with Sentry
- Backups with AWS S3 + cronjob
- Bash scripts to automate VPS setup
- Releases with GitHub Actions, GHCR, Docker

## Docs

- [`develop.md`](docs/develop.md)
- [`release.md`](docs/release.md)
- [`deploy.md`](docs/deploy.md)
- [`monitor.md`](docs/monitor.md) 
