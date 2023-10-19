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

# ... (unchanged lines)

### Production
FROM alpine:latest

# Install system dependencies
RUN apk update \
    && apk add --no-cache \
    ca-certificates \
    curl \
    tzdata \
    && update-ca-certificates

# Install MySQL client
RUN apk add --no-cache mysql-client

# Install Adminer
RUN mkdir /adminer \
    && ADMINER_URL=$(curl -s https://api.github.com/repos/vrana/adminer/releases/latest | grep "browser_download_url.*adminer.php" | cut -d : -f 2,3 | tr -d \" | tr -d ' ') \
    && curl -L $ADMINER_URL -o /adminer/index.php

# Copy executable
COPY --from=builder /app/go-chat-docker /usr/local/bin/go-chat-docker
EXPOSE 8080

# Set the working directory
WORKDIR /app

# Define environment variables for MySQL
ENV MYSQL_HOST=localhost
ENV MYSQL_PORT=3306
ENV MYSQL_USER=root
ENV MYSQL_PASSWORD=root_password
ENV MYSQL_DATABASE=my_database

# Entrypoint for running MySQL and your application
ENTRYPOINT ["/usr/local/bin/go-chat-docker"]

# Command to start MySQL in the background
CMD ["sh", "-c", "mysqld --user=mysql --datadir=/var/lib/mysql --skip-networking &"]

# Expose MySQL port
EXPOSE 3306


# Copy executable
COPY --from=builder /app/go-chat-docker /usr/local/bin/go-chat-docker
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/go-chat-docker"]