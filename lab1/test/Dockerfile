
FROM golang:1.19-alpine

WORKDIR /test

COPY go.mod ./

RUN go mod download

COPY hej.go ./

RUN go build -o /hej

CMD ["/hej"]