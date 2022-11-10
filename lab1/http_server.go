package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sync/semaphore"
)

var port int

const maxConections int64 = 10

var sem = semaphore.NewWeighted(maxConections)
var ctx context.Context

func main() {

	ctx = context.Background()

	fmt.Println(len(os.Args), os.Args)
	fmt.Println(os.Args[1])

	err := getPort()

	fmt.Println("port: ", port, err)

	http.HandleFunc("/hello", limitHandeFunc(hello))

	listen()

}

func limitHandeFunc(f http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		sem.Acquire(ctx, 1)
		defer sem.Release(1)

		f(w, req)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
	//time.Sleep(5 * time.Second)
	switch req.Method {
	case "GET":
		handleGet(w, req)
	case "POST":
		handlePost(w, req)
	default:
		http.Error(w, "501 Not Implemented", http.StatusNotImplemented)
	}
}

func getPath(w http.ResponseWriter, req *http.Request) {

}

func handleGet(w http.ResponseWriter, req *http.Request) {

	//path, format, err := getPath(w, req)

	if true {
		//http.Error(w, "404 not found.", http.StatusNotFound)
	} else {
		//http.ServeFile(w, req, path)
	}
}

func handlePost(w http.ResponseWriter, req *http.Request) {

}

func listen() {

	//log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

	l, _ := net.Listen("tcp", ":"+strconv.Itoa(port))

	// error handliing

	defer l.Close()

	host, hostPort, _ := net.SplitHostPort(l.Addr().String())

	// error handliing

	fmt.Printf("Listening on host: %s, port: %s\n", host, hostPort)

	for {

		conn, _ := l.Accept()
		fmt.Println("accpeted")
		//error handling
		sem.Acquire(ctx, 1)
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer sem.Release(1)

	request(conn)
	response(conn)

}

func request(conn net.Conn) {
	i := 0
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if i == 0 {
			m := strings.Fields(line)[0]
			fmt.Println("Methods", m)
		}
		if line == "" {
			break
		}
		i++
	}

}
func response(conn net.Conn) {
	body := "This Is Go Http Server Using TCP"

	fmt.Fprint(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	fmt.Fprint(conn, "\r\n")
	fmt.Fprint(conn, body)
}

func getPort() bool {

	sPort := os.Args[1]
	portNum, err := strconv.Atoi(sPort)

	if err != nil {
		return false
	}

	port = portNum
	return true
}
