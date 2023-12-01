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
COPY ./gorillachat /app
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

# Install phpMyAdmin
RUN apk add --no-cache phpmyadmin

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

# Copy the startup script
COPY startup.sh /usr/local/bin/startup.sh

# Make the script executable
RUN chmod +x /usr/local/bin/startup.sh

# Expose MySQL port
EXPOSE 3306

# Expose phpMyAdmin port
EXPOSE 8081

# Run the startup script
CMD ["/usr/local/bin/startup.sh"]

# Copy executable
COPY --from=builder /app/go-chat-docker /usr/local/bin/go-chat-docker
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/go-chat-docker"]