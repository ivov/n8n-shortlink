# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /builder-dir
COPY go.mod go.sum ./ 
RUN go mod download
COPY . ./
ENV CGO_ENABLED=1
RUN apk add --no-cache git build-base sqlite 
RUN go mod download
RUN BUILD_TIME=$(date +"%Y-%m-%dT%H:%M:%S%z") && \
   GIT_COMMIT=$(git rev-parse --short HEAD) && \
   go build -ldflags="-s -X main.commitSha=$GIT_COMMIT -X main.buildTime=$BUILD_TIME" -o bin cmd/server/main.go

FROM alpine:latest
RUN mkdir /root/.n8n-shortlink
WORKDIR /root/n8n-shortlink
COPY --from=builder /builder-dir/bin bin
COPY --from=builder /builder-dir/internal/db/migrations internal/db/migrations
EXPOSE 3001
CMD ["./bin/main"]
