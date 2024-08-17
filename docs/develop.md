# Development

Install Go 1.22.5:

```sh
brew install go@1.22.5
```

Install Go tooling:

```sh
go install gotest.tools/gotestsum@latest
go install golang.org/x/lint/golint@latest
go install github.com/air-verse/air@latest
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Clone repository:

```sh
git clone git@github.com:ivov/n8n-shortlink.git && cd n8n-shortlink
```

Create an alias:

```sh
echo "alias s.go='cd $(pwd)'" >> ~/.zshrc && source ~/.zshrc
```

Refer to the [Makefile](../Makefile):

```sh
make help
```

> [!IMPORTANT]  
> HTML files are embedded in the binary at compile time, so if you change them, you need to rebuild the binary, else use `make live` to start the server with live reload.

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