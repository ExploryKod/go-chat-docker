FROM golang:1.21-alpine AS base
WORKDIR /app

ENV GO111MODULE="on"
ENV GOOS="linux"
ENV GOARCH=amd64
ENV CGO_ENABLED=0

# System dependencies
RUN apk update \
    && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

### Development with hot reload and debugger
FROM base AS dev
WORKDIR /app

# Hot reloading mod
RUN go install github.com/cosmtrek/air@latest

EXPOSE 8080
EXPOSE 2345
RUN air init
ENTRYPOINT ["air"]

### Executable builder
FROM base AS builder
WORKDIR /app

# Application dependencies
COPY . /app
RUN go mod download \
    && go mod verify

RUN go build -o go-chat-docker -a .

### Production
FROM alpine:latest

RUN apk update \
    && apk add --no-cache \
    ca-certificates \
    curl \
    tzdata \
    && update-ca-certificates

# Copy executable
COPY --from=builder /app/go-chat-docker /usr/local/bin/go-chat-docker
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/go-chat-docker"]