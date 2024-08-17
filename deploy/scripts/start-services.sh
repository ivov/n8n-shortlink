#!/bin/bash

# Script to start server and monitoring stack: deploy/scripts/start-server.sh

GITHUB_USER=$(grep GITHUB_USER $HOME/deploy/.config | cut -d'=' -f2 | tr -d '"*')
GITHUB_REPO=$(grep GITHUB_REPO $HOME/deploy/.config | cut -d'=' -f2 | tr -d '"*')

RAW_REPO_ROOT="https://raw.githubusercontent.com/$GITHUB_USER/$GITHUB_REPO"

# update docker-compose files
curl -o deploy/docker-compose.monitoring.yml $RAW_REPO_ROOT/main/deploy/docker-compose.monitoring.yml
curl -o deploy/docker-compose.yml $RAW_REPO_ROOT/main/deploy/docker-compose.yml

docker-compose --file deploy/docker-compose.monitoring.yml pull
docker-compose --file deploy/docker-compose.yml pull

COMPOSE_PROJECT_NAME=n8n_shortlink docker-compose --file deploy/docker-compose.monitoring.yml up -d
COMPOSE_PROJECT_NAME=n8n_shortlink docker-compose --file deploy/docker-compose.yml --profile production up -d

# Adding `COMPOSE_PROJECT_NAME=n8n_shortlink` prevents Docker from prefixing volume names with this dir name.