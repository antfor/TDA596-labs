# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-alpine

WORKDIR /lab1

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY http_server.go ./

RUN go build -o /http_server

COPY data/ ./data

## Deploy 

#COPY --from=build /http_server /http_server

#EXPOSE 8000

CMD ["/http_server", "8000"]