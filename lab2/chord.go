package main

import (
	"bufio"
	"bytes"
	"context"
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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var me *node.Node

const serverPort = "80"

func main() {

	arg, argType := argument.GetArg()

	fmt.Println("args:", arg, "type: ", argType)

	id := node.Hash(arg.A + ":" + strconv.Itoa(arg.P))

	if arg.I != "" {
		id = node.Mod(arg.I)
	}

	me = &node.Node{Ip: arg.A, Port: arg.P, Id: id}
	RPC_server(me)

	if argType == argument.Create {

		err := me.Create(&arg.R, &node.Empty{})

		fmt.Println("create err:", err)

	} else if argType == argument.Join {

		Jnode := &node.Node{}
		err := node.GetNode(arg.Ja, arg.Jp, Jnode)
		if err != nil {
			panic(err)
		}

		err = me.Join(Jnode, &node.Empty{})

		fmt.Println("join err:", err)

		from, files := me.TakesKeys()

		moveFiles(from, me, files)

	} else {
		panic("invalid arguments")
	}

	go timer(arg.Tcp, func() { me.Check_predessesor(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Ts, func() { me.Stabilze(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Tff, func() { me.Fix_fingers(&node.Empty{}, &node.Empty{}) })

	//CreateNewContainer("http_server")

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
			me.PrintState()
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

func CreateNewContainer(image string) (string, error) {
	port := "80" //strconv.Itoa(me.Port)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: port,
	}
	containerPort, err := nat.NewPort("tcp", "80")
	if err != nil {
		panic("Unable to get the port")
	}

	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cont, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: image,
		},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	fmt.Printf("Container %s is started", cont.ID)
	return cont.ID, nil
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

		response, err := http.Get(fromUrl)

		if err != nil {
			fmt.Println("error in moveFiles (Get): ", err)
		}

		httpDelete(fromUrl)

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
