# Develop

This guide explains how to set up a local development environment.

1. Install tooling:

```sh
brew install go@1.23.3
go install gotest.tools/gotestsum@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
brew install shellcheck@0.10.0
```

2. Clone repository:

```sh
git clone git@github.com:ivov/n8n-shortlink.git
```

3. Create a DB at `~/.n8n-shortlink/n8n-shortlink.sqlite` and run migrations:

```sh
cd n8n-shortlink
make db/create
make db/mig/up
```

For all `Makefile` commands:

```sh
make help
```

> [!IMPORTANT]  
> When developing locally, keep in mind that HTML files are embedded in the binary at compile time, so if you change HTML files, you need to rebuild the binary, else use `make live` to start the server with live reload.

## Sample requests

Sample requests to create URL shortlink:

```sh
curl -X POST http://localhost:3001/shortlink -d '{ "content": "https://ivov.dev" }'
curl -X POST http://localhost:3001/shortlink -d '{ "content": "https://ivov.dev", "slug": "my-url" }'
```

Sample request to create workflow shortlink:

```sh
curl -X POST http://localhost:3001/shortlink -H "Content-Type: application/json" -d '{ "slug": "my-workflow", "content": "{\"nodes\":[{\"parameters\":{},\"id\":\"f6c01408-2371-4542-b4fa-abbfa61b0ef2\",\"name\":\"When clicking \u2018Test workflow\u2019\",\"type\":\"n8n-nodes-base.manualTrigger\",\"typeVersion\":1,\"position\":[580,300]},{\"parameters\":{\"options\":{}},\"id\":\"0cf6ba0e-b33e-4a8d-9dd0-10f4fdcc42c2\",\"name\":\"Edit Fields\",\"type\":\"n8n-nodes-base.set\",\"typeVersion\":3.4,\"position\":[800,300]}],\"connections\":{\"When clicking \u2018Test workflow\u2019\":{\"main\":[[{\"node\":\"Edit Fields\",\"type\":\"main\",\"index\":0}]]}},\"pinData\":{}}" }'
```

Sample requests to resolve shortlinks:

```sh
curl http://localhost:3001/my-url
curl http://localhost:3001/my-workflow
curl http://localhost:3001/my-workflow/view
```

Sample requests for health and metrics:

```sh
curl http://localhost:3001/health
curl http://localhost:3001/metrics
curl http://localhost:3001/debug/vars
```
