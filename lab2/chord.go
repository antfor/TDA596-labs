package main

import (
	"bufio"
	"bytes"
	"fmt"
	"lab2/argument"
	"lab2/node"
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

const serverPort = "80"

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
		cmd, arg := parseCmd(input)

		switch cmd {
		case "PrintState":
			me.PrintState()
		case "Lookup":
			lookup(arg)
		case "StoreFile":
			storeFile(arg)
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

func storeFile(file string) {
	key := node.Hash(file)
	reply := &node.Node{}
	me.Find_successor(&key, reply)

	fileName := filepath.Base(file)

	body, err := os.ReadFile(file)

	if err == nil {

		content := http.DetectContentType(body)

		reader := bytes.NewReader(body)

		fmt.Println("id is: ", reply.Id)
		//response, _ := http.Post("http://"+reply.Ip+":"+strconv.Itoa(reply.Port)+"/"+fileName, content, reader)
		response, _ := http.Post("http://"+reply.Ip+":"+serverPort+"/"+fileName, content, reader)

		reply.StoreFile(key, file)

		fmt.Println(response)

	} else {
		fmt.Println("error reading file: ", err)
	}

}

func parseCmd(input string) (string, string) {
	var cmd, arg string
	inputList := strings.Split(input, " ")

	if 0 < len(inputList) {
		cmd = inputList[0]
	}
	if 1 < len(inputList) {
		arg = inputList[1]
	}

	arg = strings.Replace(arg, "\n", "", -1)
	arg = strings.Replace(arg, "\r", "", -1)

	cmd = strings.Replace(cmd, "\n", "", -1)
	cmd = strings.Replace(cmd, "\r", "", -1)

	return cmd, arg
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

		fromUrl := "http://" + from.Ip + ":" + serverPort + "/" + file // todo: change from serverPort

		//Get
		response, err := http.Get(fromUrl)

		if err != nil {
			fmt.Println("error in moveFiles (Get): ", err)
		}

		//Delete
		httpDelete(fromUrl)

		//Post
		_, err = http.Post("http://"+to.Ip+":"+serverPort+"/"+file, response.Header.Get("Content-Type"), response.Body) // todo: change from serverPort

		if err != nil {
			fmt.Println("error in moveFiles (Post): ", err)
		}

	}
}

func httpDelete(url string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err

	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func createNode(arg argument.Argument) *node.Node {

	id := node.Hash(arg.A + ":" + strconv.Itoa(arg.P))

	if arg.I != "" {
		id = node.Mod(arg.I)
	}

	return &node.Node{Ip: arg.A, Port: arg.P, Id: id, R: arg.R}

}
