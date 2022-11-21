package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strconv"

	"golang.org/x/sync/semaphore"
)

const maxConections int64 = 10

var sem = semaphore.NewWeighted(maxConections)
var ctx context.Context

func main() {

	ctx = context.Background()

	port, valid := getPort()

	if !valid {
		panic("not a valid port")
	}

	listen(port)
}

func listen(port int) {

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		panic("could not listen to port")
	}

	defer l.Close()

	for {

		conn, err := l.Accept()

		if err != nil {
			panic("error in accpeting connection")
		}

		fmt.Println("accpeted connection")
		fmt.Println(runtime.NumGoroutine())

		sem.Acquire(ctx, 1)

		go func() {

			defer conn.Close()
			defer sem.Release(1)

			fmt.Println("serve connection")
			handleConnection(conn)

			fmt.Println("done")

		}()

	}

}

func handleConnection(conn net.Conn) {

	rw := httptest.NewRecorder()
	req, err := request(conn)
	msgClient := false

	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	} else {
		msgClient = handleRequest(conn, req, rw)
	}

	if !msgClient {
		rw.Result().Write(conn)
	}

}

func request(conn net.Conn) (*http.Request, error) {
	scanner := bufio.NewReader(conn)
	return http.ReadRequest(scanner)
}

func handleRequest(conn net.Conn, req *http.Request, rw http.ResponseWriter) bool {

	switch req.Method {
	case http.MethodGet: //GET
		return handleGet(conn, req, rw)
	default:
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return false
	}
}

func handleGet(conn net.Conn, req *http.Request, rw http.ResponseWriter) bool {

	url := req.URL.String()

	if len(url) > 0 && url[0] == '/' {
		url = url[1:]
	}

	resp, err := http.Get(url)

	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return false
	} else {
		resp.Write(conn)
		return true
	}

}

func getPort() (int, bool) {

	sPort := os.Args[1]
	portNum, err := strconv.Atoi(sPort)

	if err != nil || portNum < 0 || portNum > 65535 {
		return -1, false
	}

	return portNum, true
}
