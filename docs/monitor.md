# Monitor

## Overview

In the `n8n-shortlink-network` network, we have these services running:

- http://n8n-shortlink:3001 → Shortlink API exposes Prometheus metrics, some derived from [expvar](https://pkg.go.dev/expvar).  
- http://prometheus:9090 → [Prometheus](https://prometheus.io/) to scrape and aggregate metrics.
- http://node_exporter:9100 → [Node Exporter](https://github.com/prometheus/node_exporter) to expose host metrics.
- http://cadvisor:8080 → [cAdvisor](https://github.com/google/cadvisor) to expose container metrics.
- http://loki:3100 → [Loki](https://grafana.com/oss/loki/) to aggregate logs from Promtail.
- [Promtail](https://grafana.com/docs/loki/latest/send-data/promtail/) (no port exposed) to push logs to Loki.
- http://grafana:3000 → [Grafana](https://grafana.com/) to visualize metrics and logs.

Prometheus aggregates metrics from itself, the API, node exporter, cAdvisor. Loki aggregates logs from Promtail, which pushes them from the API and from the backup cronjob. Grafana visualizes metrics and logs at `grafana.n8n.to`. Outside the network, e.g. to the reverse proxy, these services are available as `localhost:{port}`.

## TODOs

- [ ] Set up Grafana dashboards
- [ ] Set up UptimeRobot or Checkly
- [ ] Set up https://healthchecks.io/ for backup cronjob
- [ ] Monitor sqlite DB (how?)
- [ ] Makefile commands for server logs
- [ ] Makefile commands for backups
- [ ] Makefile command for Sentry
