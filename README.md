# n8n-shortlink

Golang app for creating and resolving shortlinks for n8n workflows and URLs.

Small learning project to get familiar with deployment best practices.

Features:

- Create + resolve shortlinks for n8n workflows and URLs
- Optionally render n8n workflow shortlinks on canvas
- Vanity URLs and password protection supported
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
- Uptime monitoring with UptimeRobot
- Releases with GitHub Actions, GHCR, Docker

## Author

© 2024 Iván Ovejero
