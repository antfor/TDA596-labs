package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sync/semaphore"
)

const maxConections int64 = 10

var sem = semaphore.NewWeighted(maxConections)
var ctx context.Context

const baseDir string = "./data"

var validExt = [...]string{"html", "txt", "gif", "jpeg", "jpg", "css"}
var validContent = [...]string{"text/html", "text/plain", "image/gif", "image/jpeg", "image/jpeg", "text/css"}
var mapContentToExt = func() map[string]string {
	x := make(map[string]string)
	for i := 0; i < len(validExt); i++ {
		x[validContent[i]] = validExt[i]
	}
	return x
}()

func main() {

	ctx = context.Background()

	port, err := getPort()

	if !err {
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

		go func() {
			sem.Acquire(ctx, 1)
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

	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	} else {
		handleRequest(conn, req, rw)
	}

	rw.Result().Write(conn)
}

func request(conn net.Conn) (*http.Request, error) {
	scanner := bufio.NewReader(conn)
	return http.ReadRequest(scanner)
}

func handleRequest(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	switch req.Method {
	case http.MethodGet: //GET
		handleGet(conn, req, rw)
	case http.MethodPost: // POST
		handlePost(conn, req, rw)
	default:
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

func handleGet(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	ext := filepath.Ext(req.URL.Path)

	if len(ext) > 0 {
		ext = ext[1:]
	}

	if validExtension(ext) {
		http.ServeFile(rw, req, baseDir+req.URL.Path)
	} else {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func handlePost(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	content := toContent(http.DetectContentType(body))
	ext := mapContentToExt[content]

	path := baseDir + req.URL.Path
	pathExt := filepath.Ext(req.URL.Path)

	if len(ext) > 0 {
		pathExt = pathExt[1:]
	}

	if validExtension(ext) && ext == pathExt {

		err := os.WriteFile(path, body, 0666)
		fmt.Println(err)
		if err != nil {
			http.Error(rw, "error writing file", http.StatusBadRequest)
		} else {
			rw.WriteHeader(http.StatusOK)
		}

	} else {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

}

func toContent(s string) string {

	return s[strings.Index(s, ":")+1 : strings.Index(s, ";")]
}

func validExtension(ext string) bool {

	for _, a := range validExt {
		if a == ext {
			return true
		}
	}
	return false
}

func getPort() (int, bool) {

	sPort := os.Args[1]
	portNum, err := strconv.Atoi(sPort)

	if err != nil {
		return -1, false
	}

	return portNum, true
}
