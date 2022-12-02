package main

import (
	"fmt"
	"lab2/argument"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type Node struct {
	ip string

	port int

	id string
}

type empty struct {
}

var predecessor *Node

var successors *Node

var fingers *Node

var me *Node

func main() {

	arg, argType := argument.NewArg()
	id := arg.A + ":" + strconv.Itoa(arg.P)

	if arg.I != "" {
		id = arg.I
	}

	me = &Node{ip: arg.A, port: arg.P, id: id}

	fmt.Println("me: ", me)
	fmt.Println("me.ip: ", me.ip, arg.A)
	fmt.Println("me.port: ", me.port, arg.P)

	go RPC_server(me)

	//time.Sleep(5000 * time.Millisecond)

	call(me, "Node.Ping", me)
	//call(me, "Ping", me)

	if argType == argument.Create {
		//	n := node{}

		//n.create()

	} else if argType == argument.Join {

		//n := node{ip: arg.Ja, port: arg.Jp}

		//n.join(n)
	} else {
		panic("invalid arguments")
	}

	for {

	}

}

func RPC_server(n *Node) {

	fmt.Println("RPC_server 1")

	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(n.port))

	if err != nil {
		log.Fatal("listen error:", err)
		fmt.Println("The error is: ", err)
	}

	for {
		fmt.Println("accept")
		l.Accept()
		fmt.Println("done")
		err := http.Serve(l, nil)
		fmt.Println("serve:")
		fmt.Println(err)
	}

}

func (n *Node) Join(np *Node, done *bool) error {

	predecessor = nil
	//	err = call(np, "func(n) {}",) //find_predecessor
	*done = true

	//	return err
	return nil
}

func (n *Node) Find_successor(id *string, nr *Node) error {
	return nil
}

func (n *Node) Closest_preceding_node(np *Node, done *bool) error {
	return nil
}

func (n *Node) Create(_ *empty, _ *empty) error {
	predecessor = nil
	successors = n
	return nil
}

func (n *Node) Stabilze(_ *empty, _ *empty) error {
	return nil
}

func (n *Node) Notify(np *Node, done *bool) error {
	return nil
}

func (n *Node) Fix_fingers(_ *empty, _ *empty) error {
	return nil
}

func (n *Node) Check_predessesor(_ *empty, _ *empty) error {

	return nil
}

func call(n *Node, f string, arg *Node) error {

	fmt.Println("call", (n.ip + ":" + strconv.Itoa(n.port)))

	//rpc.Register(n)

	client, err := rpc.DialHTTP("tcp", n.ip+":"+strconv.Itoa(n.port))

	fmt.Println("call err", err)

	if err != nil {
		return err
	}

	fmt.Println("call 2")
	err = client.Call(f, arg, &empty{})

	return err

}

func callID(n Node, f string, id string) {

}

func callEmpty(n Node, f string) {

}

func (n *Node) Ping(_ *Node, _ *empty) error {
	fmt.Println("pong")
	return nil
}
