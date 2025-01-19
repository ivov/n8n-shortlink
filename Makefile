.DEFAULT_GOAL := run

HOME_DIR := $(shell echo $$HOME)
DB_PATH := $(HOME_DIR)/.n8n-shortlink/n8n-shortlink.sqlite

CURRENT_TIME = $(shell date +"%Y-%m-%dT%H:%M:%S%z")
GIT_DESCRIPTION = $(shell git describe --always --dirty)
LINKER_FLAGS = '-s -X main.commitSha=${GIT_DESCRIPTION} -X main.buildTime=${CURRENT_TIME}'

help:
	@echo "Commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST)

# ------------
#   develop
# ------------

## run: Build and run binary
run:
	go run -ldflags=${LINKER_FLAGS} cmd/server/main.go
.PHONY: run

## live: Run binary with live reload
live:
	air
.PHONY: live

lint:
	golangci-lint run
.PHONY: lint

lintfix:
	golangci-lint run --fix
.PHONY: lintfix

# ------------
#    audit
# ------------
# sdlaksdjlask dlaksd asdlkjasdl as 
## audit: Run `go mod tidy`, `go fmt`, `golint`, `go test`, and `go vet`
audit:
	echo 'Tidying...'
	go mod tidy
	echo 'Formatting...'
	go fmt ./...
	echo 'Linting...'
	golint ./...
	echo 'Testing...'
	go test -v ./...
	echo 'Vetting...'
	go vet ./...
.PHONY: audit

## test: Run tests
test:
	gotestsum --format testname
.PHONY: test

## test/watch: Run tests in watch mode
test/watch:
	gotestsum --format testname --watch
.PHONY: test

# ------------
#    build
# ------------

## build: Build binary, burning in commit SHA and build time
build:
	go build -ldflags=${LINKER_FLAGS} -o bin cmd/server/main.go
.PHONY:build

## build/meta: Display commit SHA and build time burned into binary
build/meta:
	./bin/main -metadata-mode
.PHONY: build/meta

# ------------
#     db
# ------------

## db/create: Create new database and apply up migrations
db/create:
	@if [ -f ${DB_PATH} ]; then \
		echo "\033[0;33mWarning:\033[0m This will overwrite the existing database.\nAre you sure? (y/n)"; \
		read confirm && if [ "$$confirm" = "y" ]; then \
			rm -f ${DB_PATH}; \
			sqlite3 ${DB_PATH} ""; \
			make db/mig/up; \
			echo "\033[0;34mOK Created new empty database\033[0m"; \
		else \
			echo "\033[0;31mAborted database creation\033[0m"; \
		fi \
	else \
		sqlite3 ${DB_PATH} ""; \
		make db/mig/up; \
		echo "\033[0;34mOK Created new empty database\033[0m"; \
	fi
.PHONY: db/create

## db/connect: Connect to database
db/connect:
	sqlite3 ${DB_PATH}
.PHONY: db/connect

## db/mig/new name=$1: Create new migration, e.g. `db/mig/new name=create_users_table`
db/mig/new:
	migrate create -seq -ext=.sql -dir=./internal/db/migrations ${name}
	echo "\033[0;34mOK Created migration files\033[0m"
.PHONY: db/mig/new

## db/mig/up: Apply up migrations
db/mig/up:
	migrate -path="./internal/db/migrations" -database "sqlite3://$(DB_PATH)" up
	@echo "\033[0;34mOK Applied up migrations\033[0m"
.PHONY: db/mig/up

# ------------
#    docker
# ------------

## docker/build: Build Docker image `n8n-shortlink:local`
docker/build:
	docker build --tag n8n-shortlink:local .
.PHONY: docker/build

## docker/run: Run Docker container off image `n8n-shortlink:local`
docker/run:
	@if ! docker network inspect n8n-shortlink-network >/dev/null 2>&1; then \
		echo "Creating network n8n-shortlink-network..." && \
		docker network create n8n-shortlink-network; \
	fi && \
	docker compose --file infrastructure/03-deploy/docker-compose.yml --profile local up
.PHONY: docker/run

## docker/stop: Stop Docker container
docker/stop:
	docker stop n8n-shortlink
.PHONY: docker/run

## docker/connect: Connect to running Docker container `n8n-shortlink`
docker/connect:
	docker exec -it n8n-shortlink sh
.PHONY: docker/connect

# ------------
#     vps
# ------------

## vps/login: Log in to VPS
vps/login:
	ssh n8n-shortlink-infra
.PHONY: vps/login

## vps/logs/proxy: Tail reverse proxy logs
vps/logs/proxy:
	ssh n8n-shortlink-infra "journalctl -u caddy -f"
.PHONY: vps/logs/proxy

## vps/logs/app: Tail app container logs
vps/logs/app:
	ssh n8n-shortlink-infra "docker logs -f n8n-shortlink"
.PHONY: vps/logs/app

## vps/redeploy: Prompt watchtower to poll and deploy new version
vps/redeploy:
	ssh n8n-shortlink-infra "docker exec watchtower /watchtower --run-once n8n-shortlink"
.PHONY: vps/redeploy
