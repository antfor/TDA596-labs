package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
		w.WriteHeader(501)
	}
}

func handleGet(w http.ResponseWriter, req *http.Request) {

}

func handlePost(w http.ResponseWriter, req *http.Request) {

}

func listen() {

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
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
