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

const maxConections int64 = 10 // Max number of connections

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

/*
Listen to the given port

	For every new connection, spawn a Goroutine, at most 10

Otherwise throw a panic
*/
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

/*
Handle an accepted connection

Write the respons to the ResponseWriter NewRecorder
*/
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

/*
Get and parse the request from the connection

Return the parsed request
*/
func request(conn net.Conn) (*http.Request, error) {
	scanner := bufio.NewReader(conn)
	return http.ReadRequest(scanner)
}

/*
Call the corresponding method that was in the request

Otherwise throw an "Not implemented" error
*/
func handleRequest(conn net.Conn, req *http.Request, rw http.ResponseWriter) bool {

	switch req.Method {
	case http.MethodGet: //GET
		return handleGet(conn, req, rw)
	default:
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return false
	}
}

/*
Issue a GET to the specified URL

	Return true if it succeded and writes the result to the connection

Otherwise return false and throw a bad request error (501)
*/
func handleGet(conn net.Conn, req *http.Request, rw http.ResponseWriter) bool {

	url := req.URL.String()

	// Remove the first "/" (necessary when we are sending a proxy request from the web-browser)
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

/*
Returns the port number that the user provided as argument on the command line

	Returns the port and true if it's a valid port

Otherwise returns -1 and false
*/
func getPort() (int, bool) {

	sPort := os.Args[1]
	portNum, err := strconv.Atoi(sPort)

	if err != nil || portNum < 0 || portNum > 65535 {
		return -1, false
	}

	return portNum, true
}
