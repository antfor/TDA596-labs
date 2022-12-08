package node

import (
	"crypto/sha1"
	"fmt"
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

const m = 3 // Maybe need to change
var maxNodes = new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)

var predecessor *Node = &Node{}

var next int = 0
var fingers [m]*Node

func (n *Node) Print() {
	fmt.Println("	ip: ", n.Ip)
	fmt.Println("	port: ", n.Port)

	value, _ := strconv.ParseInt(n.Id, 16, 64)
	fmt.Println("	id (base 10)", value)
}

func (n *Node) PrintState() {
	fmt.Println("me: ")
	n.Print()
	if predecessor != nil {
		fmt.Println("predecessor: ")
		predecessor.Print()
	} else {
		fmt.Println("predecessor is nil ")
	}

	fmt.Println("successor: ")

	fingers[0].Print()
	fmt.Println("finger table: ")

	for i := 0; i < m; i++ {
		if fingers[i] != nil {
			fmt.Println("finger", i, ":")
			fingers[i].Print()
		} else {
			fmt.Println("finger ", i, " is nil")
		}
	}

}

func (n *Node) Create(_ *Empty, _ *Empty) error {
	predecessor = nil

	fingers[0] = n

	return nil
}

func (n *Node) Join(np *Node, _ *Empty) error {

	predecessor = nil

	fingers[0] = &Node{Ip: "start"}
	return Call(np, "Node.Find_successor", &n.Id, fingers[0])
}

func (n *Node) Find_successor(id *string, nr *Node) error {

	if between(*id, n.Id, fingers[0].Id, true) {

		*nr = *fingers[0]
		return nil
	} else {
		n0 := &Node{}
		err := n.Closest_preceding_node(id, n0)

		if err != nil {
			fmt.Println("error in find_suc1: ", err)
			return err
		}

		if *n0 == *n {
			*nr = *n
			return nil
		}

		return Call(n0, "Node.Find_successor", id, nr)
	}

}

func (n *Node) Closest_preceding_node(id *string, nr *Node) error {

	for i := m - 1; i >= 0; i-- {

		finger := fingers[i]

		if finger != nil {

			if between(finger.Id, n.Id, *id, false) {

				*nr = *finger
				return nil
			}
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

	err := Call(fingers[0], "Node.GetPredecessor", &Empty{}, x)

	if err != nil {
		fmt.Println("error in stab:", err)
		return err
	}

	if (*x != Node{}) {

		if between(x.Id, n.Id, fingers[0].Id, false) {
			fingers[0] = x
		}
	}

	return Call(fingers[0], "Node.Notify", n, &Empty{})
}

func (n *Node) Notify(np *Node, _ *Empty) error {

	if predecessor == nil {
		predecessor = np
		return nil
	}

	if between(np.Id, predecessor.Id, n.Id, false) {
		predecessor = np
	}
	return nil
}

func (n *Node) Fix_fingers(_ *Empty, _ *Empty) error {

	next = next + 1
	if next > m-1 {
		next = 0
	}

	id := jump(n.Id, next)

	tmp := &Node{}
	err := Call(n, "Node.Find_successor", &id, tmp)

	if err != nil {
		return err
	}

	fingers[next] = tmp
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

func (n *Node) Ping(_ *Empty, reply *string) error {

	*reply = "pong"
	return nil
}

func Call[A any, R any](n *Node, f string, arg *A, reply *R) error {

	client, err := rpc.DialHTTP("tcp", n.Ip+":"+strconv.Itoa(n.Port))

	if err != nil {
		return err
	}

	defer client.Close()
	return client.Call(f, arg, reply)
}

func Hash(elt string) string {
	hasher := sha1.New()
	hasher.Write([]byte(elt))

	hashValue := new(big.Int).SetBytes(hasher.Sum(nil))

	key := new(big.Int).Mod(hashValue, maxNodes)

	return key.Text(16)
}

func toBigInt(key string) *big.Int {
	keyID := new(big.Int)
	keyID.SetString(key, 16)

	return keyID
}

func jump(startID string, finger int) string {
	start := toBigInt(startID)
	jump := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(finger)), nil)
	dest := start.Add(start, jump)

	return new(big.Int).Mod(dest, maxNodes).Text(16)
}

// if inclusive = false: returns true if a < x < b  otherwise false
// if inclusive = true:  returns true if a < x <= b otherwise false
func between(xs string, as string, bs string, inclusive bool) bool {

	elt := toBigInt(xs)
	start := toBigInt(as)
	end := toBigInt(bs)

	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}
