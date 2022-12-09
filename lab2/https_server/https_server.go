package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

const maxConections int64 = 10 // Max number of connections

var sem = semaphore.NewWeighted(maxConections) // Cap the number of goroutines to maxConnections
var ctx context.Context

var mapFileTolock sync.Map // Map a read/write lock to every file

const baseDir string = "./data" // All files sent to the server are put in this directory

const tempDir string = "./temp"

var lock = sync.Mutex{}

var validExt = [...]string{"html", "txt", "gif", "jpeg", "jpg", "css"} // The extensions the server accepts

// Get the port, get the files and start listening
func main() {

	ctx = context.Background()

	port, valid := getPort()

	if !valid {
		panic("Not a valid port")
	}

	getFiles()

	listen(port)
}

// Iterate through all files in base directory
func getFiles() error {
	return filepath.Walk(baseDir, handleFile)
}

// Map a read write lock (pointer) to an existing file and store in a map
func handleFile(path string, info fs.FileInfo, err error) error {

	ext := filepath.Ext(path)

	if validExtension(ext) {

		var lock *sync.RWMutex = &sync.RWMutex{}

		mapFileTolock.Store(cleanURL(path), lock)

	}
	return nil
}

/*
Listen to the given port. For every new connection, spawn a Goroutine, at most 10

Otherwise throw a panic
*/
func listen(port int) {

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		panic("Could not listen to port")
	}

	defer l.Close()

	for {

		conn, err := l.Accept()

		if err != nil {
			panic("Error in accepting connection")
		}

		fmt.Println("Accepted connection")
		fmt.Println(runtime.NumGoroutine())

		sem.Acquire(ctx, 1)

		go func() {

			defer conn.Close()
			defer sem.Release(1)

			fmt.Println("Serve connection")
			handleConnection(conn)

			fmt.Println("Done")

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

	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	} else {
		handleRequest(conn, req, rw)
	}

	rw.Result().Write(conn)
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
func handleRequest(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	switch req.Method {
	case http.MethodGet: //GET
		handleGet(conn, req, rw)
	case http.MethodPost: // POST
		handlePost(conn, req, rw)
	case http.MethodDelete: // DELETE
		handleDelete(conn, req, rw)
	default:
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

/*
If the user request a valid file that exist on the server, do serverFile

Otherwise throw an appropriate error (404 or 501)

Because it's a read/write lock multiple concurrent reads can be done on the same file
*/
func handleGet(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	ext := filepath.Ext(req.URL.Path)

	if validExtension(ext) {

		path := cleanURL(baseDir + req.URL.Path)

		lockRW, found := mapFileTolock.Load(path)

		if found {
			(lockRW.(*sync.RWMutex)).RLock()

			query := req.URL.Query()

			if query.Has("pswd") { // added password query

				serveEncryptedFile(rw, req, path, query.Get("pswd"))

			} else {
				http.ServeFile(rw, req, path)
			}

			(lockRW.(*sync.RWMutex)).RUnlock()
		} else {
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}

	} else {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

/*
Write to a exising file or create a new file and store it in the base directory

	Otherwise throw an appropriate error (501)

	Because this is a read/write lock when this get hold of the lock

	No other client can read or write to it

If it's an new file it creates a new lock and add it to the sync.Map
*/
func handlePost(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	path := cleanURL(baseDir + req.URL.Path)

	pathExt := filepath.Ext(req.URL.Path)
	if len(pathExt) > 0 && pathExt[0] == '.' {
		pathExt = pathExt[1:]
	}

	if validExtension(pathExt) {

		var lock *sync.RWMutex = &sync.RWMutex{}

		lockRW, _ := mapFileTolock.LoadOrStore(path, lock)

		(lockRW.(*sync.RWMutex)).Lock()

		err := os.WriteFile(path, body, 0666)

		(lockRW.(*sync.RWMutex)).Unlock()

		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}

	} else {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

}

// Checks if the extension is in the pool of the servers accepted extensions
func validExtension(ext string) bool {

	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	for _, a := range validExt {
		if a == ext {
			return true
		}
	}
	return false
}

/*
Clean the URL, returning the shortest path name equivalent to the url

Also makes it all lowercase
*/
func cleanURL(url string) string {

	path := filepath.Clean(url)

	return strings.ToLower(path)
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

func decryptFile(url string, key string) (string, error) {

	path := cleanURL(baseDir + url)
	file, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer file.Close()
	body, err := io.ReadAll(file)

	if err != nil {
		return "", err
	}

	decrypted, err := decrypt(body, key)

	if err != nil {
		return "", err

	}

	tempPath := cleanURL(tempDir + url)

	err = os.WriteFile(tempPath, decrypted, 0666)

	if err != nil {
		return "", err
	}

	return tempPath, nil

}

func deleteFile(path string) {

	err := os.Remove(path)

	if err != nil {
		fmt.Println("error in deleteFile: ", err)
	}

}

func serveEncryptedFile(rw http.ResponseWriter, req *http.Request, path string, pswd string) {

	lock.Lock()
	decryptedPath, err := decryptFile(req.URL.Path, pswd)

	if err == nil {
		http.ServeFile(rw, req, decryptedPath)
		deleteFile(decryptedPath)

	} else {

		fmt.Println("error decrypting: ", err)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	lock.Unlock()

}

func decrypt(ciphertext []byte, key string) ([]byte, error) {
	return ciphertext, nil
}

func handleDelete(conn net.Conn, req *http.Request, rw http.ResponseWriter) {

	path := cleanURL(baseDir + req.URL.Path)

	lockRW, found := mapFileTolock.Load(path)

	if found {
		(lockRW.(*sync.RWMutex)).Lock()

		err := os.Remove(path)

		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}

		mapFileTolock.Delete(path) //copilot skrev hela functionen förutom denna raden

		(lockRW.(*sync.RWMutex)).Unlock()

	} else {
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

}
