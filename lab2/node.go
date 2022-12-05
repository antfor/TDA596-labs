package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"lab2/argument"
	"math/big"
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

const m = 32 // Maybe need to change

var predecessor *Node

var successor *Node

var fingers []*Node

var me *Node

func main() {

	arg, argType := argument.NewArg()
	id := arg.A + ":" + strconv.Itoa(arg.P)

	if arg.I != "" {
		id = arg.I
	}

	me = &Node{Ip: arg.A, Port: arg.P, Id: id}
	RPC_server(me)

	if argType == argument.Create {

		me.Create(&Empty{}, &Empty{})

	} else if argType == argument.Join {

		fmt.Println("Before: ", successor.Id)

		id := arg.Ja + ":" + strconv.Itoa(arg.Jp)
		node := &Node{Ip: arg.Ja, Port: arg.Jp, Id: id}

		me.Join(node, &Empty{})

		fmt.Println("After: ", successor.Id)

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

func (n *Node) Create(_ *Empty, _ *Empty) error {
	predecessor = nil
	successor = n
	return nil
}

func (n *Node) Join(np *Node, _ *Empty) error {

	predecessor = nil

	return call(np, "Node.Find_successor", n, successor)
}

func (n *Node) Find_successor(id *string, nr *Node) error {

	nID := hash(*id)

	if hash(n.Id).Cmp(nID) == -1 && nID.Cmp(hash(successor.Id)) >= 0 {
		nr = successor
		return nil
	} else {
		n0 := &Node{}
		err := n.Closest_preceding_node(id, n0)

		if err != nil {
			return err
		}

		return call(n0, "Node.Find_successor", id, nr)
	}

}

func (n *Node) Closest_preceding_node(id *string, nr *Node) error {

	nHash := hash(n.Id)
	idHash := hash(*id)

	for i := m; i > 0; i-- {
		fHash := hash(fingers[i].Id)

		if nHash.Cmp(fHash) == -1 && fHash.Cmp(idHash) == -1 {
			nr = fingers[i]
			return nil
		}
	}

	nr = n
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

func call[T any](n *Node, f string, arg *T, reply *Node) error {

	client, err := rpc.DialHTTP("tcp", n.Ip+":"+strconv.Itoa(n.Port))

	if err != nil {
		return err
	}

	return client.Call(f, arg, reply)
}

func (n *Node) Ping(_ *Node, _ *Node) error {
	fmt.Println("pong")
	return nil
}

func hash(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}
