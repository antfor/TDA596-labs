# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-alpine

WORKDIR /lab1_basic

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY http_server.go ./

RUN go build -o /http_server

COPY data/ ./data

EXPOSE 80

CMD ["/http_server", "80"]