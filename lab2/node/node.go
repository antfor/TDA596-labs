package node

import (
	"crypto/sha1"
	"fmt"
	"math"
	"math/big"
	"net/rpc"
	"strconv"
)

type Node struct {
	Ip string

	Port int

	Id string
}

type Empty struct {
}

const m = 4 // Maybe need to change

var predecessor *Node = &Node{}

var successor *Node = &Node{}

var next int = 0
var fingers [m]*Node

func (n *Node) Create(_ *Empty, _ *Empty) error {
	predecessor = nil
	successor = n

	for i := 0; i < m; i++ {
		fingers[i] = n
	}

	return nil
}

func (n *Node) Join(np *Node, _ *Empty) error {

	predecessor = nil

	return Call(np, "Node.Find_successor", &n.Id, successor)
}

func (n *Node) Find_successor(id *string, nr *Node) error {

	aID := toBigInt(*id)
	nID := toBigInt(n.Id)
	sID := toBigInt(successor.Id)

	if nID.Cmp(aID) == -1 && aID.Cmp(sID) >= 0 {

		*nr = *successor
		return nil
	} else {
		n0 := &Node{}
		err := n.Closest_preceding_node(id, n0)

		if err != nil {
			return err
		}

		return Call(n0, "Node.Find_successor", id, nr)
	}

}

func (n *Node) Closest_preceding_node(id *string, nr *Node) error {

	idID := toBigInt(*id)
	nID := toBigInt(n.Id)

	for i := m - 1; i >= 0; i-- {

		fID := toBigInt(fingers[i].Id)

		if nID.Cmp(fID) == -1 && fID.Cmp(idID) == -1 {
			*nr = *fingers[i]
			return nil
		}
	}

	*nr = *n
	return nil
}

func (n *Node) GetPredecessor(_ *Empty, nr *Node) error {

	if predecessor != nil {
		*nr = *predecessor
	}

	return nil
}

func (n *Node) Stabilze(_ *Empty, _ *Empty) error {

	x := &Node{}

	err := Call(successor, "Node.GetPredecessor", &Empty{}, x)

	if err != nil {
		return err
	}

	if (*x != Node{}) {
		sID := toBigInt(successor.Id)
		nID := toBigInt(n.Id)
		xID := toBigInt(x.Id)

		if nID.Cmp(xID) == -1 && xID.Cmp(sID) == -1 {
			successor = x
		}
	}

	return Call(successor, "Node.Notify", n, &Empty{})
}

func (n *Node) Notify(np *Node, _ *Empty) error {

	npID := toBigInt(np.Id)
	nID := toBigInt(n.Id)

	if predecessor == nil {
		predecessor = np
	}

	pID := toBigInt(predecessor.Id)

	if pID.Cmp(npID) == -1 && npID.Cmp(nID) == -1 {
		predecessor = np
	}
	return nil
}

func (n *Node) Fix_fingers(_ *Empty, _ *Empty) error {
	fmt.Println("fix_fingers")
	next = next + 1
	if next > m-1 {
		next = 0
	}

	nID := toBigInt(n.Id)
	pow := big.NewInt(int64(math.Pow(2, float64(next-1))))
	id := nID.Add(nID, pow).Text(16)

	Call(n, "Node.Find_successor", &id, fingers[next])
	return nil
}

func (n *Node) Check_predessesor(_ *Empty, _ *Empty) error {

	if predecessor != nil {
		var reply string

		err := Call(predecessor, "Node.Ping", &Empty{}, &reply)

		if err != nil && reply != "pong" {
			predecessor = nil
			return err
		}
	}
	return nil
}

func Call[A any, R any](n *Node, f string, arg *A, reply *R) error {

	client, err := rpc.DialHTTP("tcp", n.Ip+":"+strconv.Itoa(n.Port))

	if err != nil {
		return err
	}

	return client.Call(f, arg, reply)
}

func Hash(elt string) string {
	hasher := sha1.New()
	hasher.Write([]byte(elt))

	hashValue := new(big.Int).SetBytes(hasher.Sum(nil))

	key := new(big.Int).Mod(hashValue, big.NewInt(int64(math.Pow(2, m))))
	return key.Text(16)
}

func toBigInt(key string) *big.Int {
	keyID := new(big.Int)
	keyID.SetString(key, 16)

	return keyID
}

func (n *Node) Ping(_ *Empty, reply *string) error {

	*reply = "pong"
	return nil
}
