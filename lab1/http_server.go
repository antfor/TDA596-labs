package main

import (
	"bufio"
	"context"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sync/semaphore"
)

var port int

const maxConections int64 = 10

var sem = semaphore.NewWeighted(maxConections)
var ctx context.Context

var validExt = [...]string{"html", "txt", "gif", "jpeg", "jpg", "css"}

func main() {

	ctx = context.Background()

	fmt.Println(len(os.Args), os.Args)
	fmt.Println(os.Args[1])

	err := getPort()

	fmt.Println("port: ", port, err)

	listen()

}

func limitHandeFunc(f http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		sem.Acquire(ctx, 1)
		defer sem.Release(1)

		f(w, req)
	}
}

func handleRequest(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	switch req.Method {
	case http.MethodGet: //GET
		handleGet(conn, req, rw)
	case http.MethodPost: // POST
		handlePost(conn, req, rw)
	default:
		//http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

func getPath(conn net.Conn, req *http.Request) {

}

func handleGet(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	ext := filepath.Ext(req.URL.Path)[1:]
	mimeType := mime.TypeByExtension(ext)
	dat, err := os.ReadFile("." + req.URL.Path)

	if err == nil {
		//http.Error(w, "404 not found.", http.StatusNotFound)
	} else if validExtension(ext) {

		if validExtension(ext) {
			//error
		}

		fmt.Printf(ext)
		fmt.Println(mimeType)

	}
}

func validExtension(ext string) bool {

	for _, a := range validExt {
		if a == ext {
			return true
		}
	}
	return false
}

func handlePost(conn net.Conn, req *http.Request) {
	//http.DetectContentType()
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

	req, _ := request(conn)
	//todo error handling
	rw := httptest.NewRecorder()
	handleRequest(conn, req, rw)
	response(conn, rw.Result())

}

func request(conn net.Conn) (req *http.Request, err error) {

	scanner := bufio.NewReader(conn)
	req, err = http.ReadRequest(scanner)

	if err != nil {

	}

	fmt.Println(req)
	fmt.Println(err)
	fmt.Println(req.URL)
	fmt.Println(req.URL.Path)

	return
}

func response(conn net.Conn, r *http.Response) {

	r.Write(conn)

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
