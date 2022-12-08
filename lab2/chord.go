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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var me *node.Node

func main() {

	arg, argType := argument.GetArg()

	fmt.Println("args:", arg, "type: ", argType)

	id := node.Hash(arg.A + ":" + strconv.Itoa(arg.P))

	if arg.I != "" {
		id = node.Hash(arg.I)
	}

	me = &node.Node{Ip: arg.A, Port: arg.P, Id: id}
	RPC_server(me)

	if argType == argument.Create {

		err := me.Create(&node.Empty{}, &node.Empty{})

		fmt.Println("create err:", err)

	} else if argType == argument.Join {

		id := node.Hash(arg.Ja + ":" + strconv.Itoa(arg.Jp)) // todo call getId

		Jnode := &node.Node{Ip: arg.Ja, Port: arg.Jp, Id: id}

		err := me.Join(Jnode, &node.Empty{})

		fmt.Println("join err:", err)

	} else {
		panic("invalid arguments")
	}

	go timer(arg.Tcp, func() { me.Check_predessesor(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Ts, func() { me.Stabilze(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Tff, func() { me.Fix_fingers(&node.Empty{}, &node.Empty{}) })

	startServer()

	read_stdin()

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
		}

		//fmt.Println(runtime.NumGoroutine())
		//me.PrintState()
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

func startServer() { // todo do only if docker is not running
	port := strconv.Itoa(me.Port)
	cmd2 := exec.Command("docker", "images")
	var out bytes.Buffer
	cmd2.Stdout = &out

	cmd2.Run()

	fmt.Printf("translated phrase: %q\n", out.String())

	fmt.Println(out)

	cmd := exec.Command("docker", "run", "--publish "+port+":80", "http_server") // todo not on the cloud

	err := cmd.Start()

	fmt.Println("server:", err)
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
		//http.Post("http://"+reply.Ip+":"+strconv.Itoa(reply.Port)+"/"+fileName, content, reader)

		response, err2 := http.Post("http://"+reply.Ip+":80"+"/"+fileName, content, reader)

		fmt.Println(response, err2)
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
