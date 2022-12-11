package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"io"
	"lab2/argument"
	"lab2/node"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var me *node.Node

const serverPort = "443"

func main() {

	arg, argType := argument.GetArg()

	fmt.Println("args:", arg, "type: ", argType)

	me = createNode(arg)
	RPC_server(me)

	switch argType {
	case argument.Create:
		create(arg)
	case argument.Join:
		join(arg)
	case argument.NotValid:
		panic("invalid arguments")
	}

	go timer(arg.Tcp, func() { me.Check_predessesor(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Ts, func() { me.Stabilze(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Tff, func() { me.Fix_fingers(&node.Empty{}, &node.Empty{}) })

	read_stdin()

}

func create(arg argument.Argument) {
	err := me.Create(&arg.R, &node.Empty{})

	if err != nil {
		fmt.Println("create err:", err)
	}

}

func join(arg argument.Argument) {
	Jnode := &node.Node{}
	err := node.GetNode(arg.Ja, arg.Jp, Jnode)
	if err != nil {
		panic(err)
	}

	err = me.Join(Jnode, &node.Empty{})

	if err != nil {
		fmt.Println("join err:", err)
	}

	from, files := me.TakesKeys()

	moveFiles(from, me, files)
}

func read_stdin() {

	reader := bufio.NewReader(os.Stdin)

	for {
		input, _ := reader.ReadString('\n')
		cmd, arg, option := parseCmd(input)

		switch cmd {
		case "PrintState":
			me.PrintState()
		case "Lookup":
			lookup(arg)
		case "StoreFile":
			storeFile(arg, option)
		default:
			fmt.Println("Not a valid command")
			me.PrintState()
		}

	}
}

func lookup(file string) {

	key := node.Hash(file)
	reply := &node.Node{}
	me.Find_successor(&key, reply)

	fmt.Println("Key hashed is: ", key)

	fmt.Println("File is at node: ")
	reply.Print()
}

func storeFile(file string, option string) {
	fmt.Println("storing file: ", file, " with option: ", option)
	key := node.Hash(file)
	reply := &node.Node{}
	me.Find_successor(&key, reply)

	fileName := filepath.Base(file)

	body, err := os.ReadFile(file)

	if err == nil {

		content := http.DetectContentType(body)

		if option != "" {
			fmt.Println("encrypting file with key: ", option)
			body, err = Encrypt(body, option)

			if err != nil {
				fmt.Println("error encrypting file: ", err)
			}
		}

		reader := io.NopCloser(bytes.NewReader(body))

		fmt.Println("id is: ", reply.Id)

		//POST
		//response, _ := http.Post("http://"+reply.Ip+":"+serverPort+"/"+fileName, content, reader)
		err = httpsPost(reply.Ip+":"+serverPort, "https://"+reply.Ip+":"+serverPort+"/"+fileName, content, reader)

		if err != nil {
			fmt.Println("error posting file: ", err)
		}
		reply.StoreFile(key, file)

	} else {
		fmt.Println("error reading file: ", err)
	}

}

func parseCmd(input string) (string, string, string) {
	var cmd, arg, option string
	inputList := strings.Split(input, " ")

	if 0 < len(inputList) {
		cmd = inputList[0]
	}
	if 1 < len(inputList) {
		arg = inputList[1]
	}
	if 2 < len(inputList) {
		option = inputList[2]
	}

	arg = strings.Replace(arg, "\n", "", -1)
	arg = strings.Replace(arg, "\r", "", -1)

	cmd = strings.Replace(cmd, "\n", "", -1)
	cmd = strings.Replace(cmd, "\r", "", -1)

	option = strings.Replace(option, "\n", "", -1)
	option = strings.Replace(option, "\r", "", -1)

	return cmd, arg, option
}

func timer(timeMs int, f func()) {
	for {
		time.Sleep(time.Duration(timeMs) * time.Millisecond)
		f()
	}
}

func RPC_server(n *node.Node) {

	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(n.Port))

	if err != nil {
		fmt.Println("error in RPC_server: ", err)
	}
	go http.Serve(l, nil)
}

func moveFiles(from *node.Node, to *node.Node, files []string) {

	fmt.Println("moveFiles: ", files)
	fmt.Println("from: ", from)
	fmt.Println("to: ", to)
	for _, file := range files {

		fmt.Println("file: ", file)

		fromUrl := "https://" + from.Ip + ":" + serverPort + "/" + file // todo: change from serverPort

		//Get
		//	response, err := http.Get(fromUrl)
		response, err := httpsGet(from.Ip+":"+serverPort, fromUrl)

		if err != nil {
			fmt.Println("error in moveFiles (Get): ", err)
		}

		//Delete
		//httpDelete(fromUrl)
		err = httpsDelete(from.Ip, fromUrl)
		if err != nil {
			fmt.Println("error in moveFiles (DELETE): ", err)
		}

		//Post
		//_, err = http.Post("http://"+to.Ip+":"+serverPort+"/"+file, response.Header.Get("Content-Type"), response.Body) // todo: change from serverPort
		err = httpsPost(to.Ip+":"+serverPort, "https://"+to.Ip+":"+serverPort+"/"+file, response.Header.Get("Content-Type"), response.Body)

		if err != nil {
			fmt.Println("error in moveFiles (Post): ", err)
		}

	}
}

func httpsGet(ip string, url string) (*http.Response, error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", ip, conf)
	if err != nil {
		fmt.Println("error in https(dial): ", err)
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error in https(newRequest): ", err)
		return nil, err
	}

	err = req.Write(conn)

	defer conn.Close()

	res, err := http.ReadResponse(bufio.NewReader(conn), req)

	return res, nil

}

func httpsPost(ip string, url string, content string, body io.ReadCloser) error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", ip, conf)
	if err != nil {
		fmt.Println("error in https(dial): ", err)
		return err
	}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("error in https(newRequest): ", err)
	}

	req.Body = body
	req.Header.Set("Content-Type", content)

	err = req.Write(conn)

	defer conn.Close()

	return nil
}

func httpsDelete(ip string, url string) error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", ip, conf)
	if err != nil {
		fmt.Println("error in https(dial): ", err)
		return err
	}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("error in https(newRequest): ", err)
		return err
	}

	err = req.Write(conn)

	defer conn.Close()

	return nil

}

func createNode(arg argument.Argument) *node.Node {

	id := node.Hash(arg.A + ":" + strconv.Itoa(arg.P))

	if arg.I != "" {
		id = node.Mod(arg.I)
	}

	return &node.Node{Ip: arg.A, Port: arg.P, Id: id, R: arg.R}

}

// Encryptins from: https://blog.logrocket.com/learn-golang-encryption-decryption/

var rand_bytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(plainText []byte, MySecret string) ([]byte, error) {

	key := genKey(MySecret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return plainText, err
	}

	cfb := cipher.NewCFBEncrypter(block, rand_bytes)
	cipherText := make([]byte, len(plainText))

	cfb.XORKeyStream(cipherText, plainText)
	return cipherText, nil
}

func genKey(pswd string) []byte {

	hasher := sha1.New()
	hasher.Write([]byte(pswd))

	hashValue := new(big.Int).SetBytes(hasher.Sum(nil))

	return hashValue.Bytes()[:16]
}
