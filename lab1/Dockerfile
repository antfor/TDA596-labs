# syntax=docker/dockerfile:1

## Build
FROM golang:1.16-alpine

WORKDIR /lab1

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY http_server.go ./

RUN go build -o /http_server.go


## Deploy 

#COPY --from=build /http_server /http_server

EXPOSE 8000

CMD ["/http_server"]