# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-alpine

WORKDIR /lab1_bonus

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY proxy.go ./

RUN go build -o /proxy

## Deploy 

EXPOSE 80

CMD ["/proxy", "80"]