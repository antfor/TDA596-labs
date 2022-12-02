package main

import (
	"bufio"
	"fmt"
	"lab2/argument"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

type Node struct {
	Ip string

	Port int

	Id string
}

type Empty struct {
}

var predecessor *Node

var successor *Node

var fingers *Node

var me *Node

func main() {

	arg, argType := argument.NewArg()
	id := arg.A + ":" + strconv.Itoa(arg.P)

	if arg.I != "" {
		id = arg.I
	}

	me = &Node{Ip: arg.A, Port: arg.P, Id: id}

	if argType == argument.Create {

		fmt.Println("RPC")
		RPC_server(me)

	} else if argType == argument.Join {

		call(me, "Node.Ping", me, me)

	} else {
		panic("invalid arguments")
	}

	read_stdin()

}

func read_stdin() {

	reader := bufio.NewReader(os.Stdin)

	for {
		cmd, _ := reader.ReadString('\n')
		fmt.Println(cmd)
	}
}

func RPC_server(n *Node) {

	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(n.Port))

	if err != nil {
		fmt.Println("error in RPC_server: ", err)
	}
	go http.Serve(l, nil)
}

func (n *Node) Join(np *Node, _ *Empty) error {

	predecessor = nil

	return call(np, "Node.Find_successor", n, successor)
}

func (n *Node) Find_successor(id *string, nr *Node) error {
	return nil
}

func (n *Node) Closest_preceding_node(np *Node, done *bool) error {
	return nil
}

func (n *Node) Create(_ *Empty, _ *Empty) error {
	predecessor = nil
	successor = n
	return nil
}

func (n *Node) Stabilze(_ *Empty, _ *Empty) error {
	return nil
}

func (n *Node) Notify(np *Node, done *bool) error {
	return nil
}

func (n *Node) Fix_fingers(_ *Empty, _ *Empty) error {
	return nil
}

func (n *Node) Check_predessesor(_ *Empty, _ *Empty) error {

	return nil
}

func call(n *Node, f string, arg *Node, reply *Node) error {

	client, err := rpc.DialHTTP("tcp", n.Ip+":"+strconv.Itoa(n.Port))

	if err != nil {
		return err
	}

	return client.Call(f, arg, reply)
}

func callID(n Node, f string, id string) {

}

func callEmpty(n Node, f string) {

}

func (n *Node) Ping(_ *Node, _ *Node) error {
	fmt.Println("pong")
	return nil
}
