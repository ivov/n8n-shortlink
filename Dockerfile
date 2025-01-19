# syntax=docker/dockerfile:1

FROM --platform=${TARGETPLATFORM} golang:1.23-alpine AS builder
WORKDIR /builder-dir
COPY . ./
ENV CGO_ENABLED=1
RUN apk add --no-cache git build-base sqlite 
RUN go mod download
RUN go build -o bin cmd/server/main.go

FROM --platform=${TARGETPLATFORM} alpine:3.21.2
RUN mkdir /root/.n8n-shortlink
WORKDIR /root/n8n-shortlink
COPY --from=builder /builder-dir/bin bin
COPY --from=builder /builder-dir/internal/db/migrations internal/db/migrations
EXPOSE 3001
CMD ["./bin/main"]

