package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"lab2/argument"
	"math"
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

var predecessor *Node = &Node{}

var successor *Node = &Node{}

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

		id := arg.Ja + ":" + strconv.Itoa(arg.Jp)
		node := &Node{Ip: arg.Ja, Port: arg.Jp, Id: id}

		me.Join(node, &Empty{})

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

	return call(np, "Node.Find_successor", &n.Id, successor)
}

func (n *Node) Find_successor(id *string, nr *Node) error {

	nID := hash(*id)

	if hash(n.Id).Cmp(nID) == -1 && nID.Cmp(hash(successor.Id)) >= 0 {
		fmt.Println("hello?", successor.Id)
		*nr = *successor
		fmt.Println("hello2?", nr.Id)
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
			*nr = *fingers[i]
			return nil
		}
	}

	*nr = *n
	return nil
}

func (n *Node) GetPredecessor(_ *Empty, nr *Node) error {
	*nr = *predecessor
	return nil
}

func (n *Node) Stabilze(_ *Empty, _ *Empty) error {

	x := &Node{}

	err := call(successor, "Node.GetPredecessor", &Empty{}, x)

	if err != nil {
		return err
	}

	nHash := hash(n.Id)
	sHash := hash(successor.Id)
	xHash := hash(x.Id)

	if nHash.Cmp(xHash) == -1 && xHash.Cmp(sHash) == -1 {
		successor = x
	}

	return call(successor, "Node.Notify", n, &Empty{})
}

func (n *Node) Notify(np *Node, _ *Empty) error {

	npHash := hash(np.Id)
	nHash := hash(n.Id)

	if predecessor == nil || (hash(predecessor.Id).Cmp(npHash) == -1 && npHash.Cmp(nHash) == -1) {
		predecessor = np
	}
	return nil
}

func (n *Node) Fix_fingers(_ *Empty, _ *Empty) error {
	return nil
}

func (n *Node) Check_predessesor(_ *Empty, _ *Empty) error {

	return nil
}

func call[A any, R any](n *Node, f string, arg *A, reply *R) error {

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

	hashValue := new(big.Int).SetBytes(hasher.Sum(nil))

	key := new(big.Int).Mod(hashValue, big.NewInt(int64(math.Pow(2, m))))
	return key
}
