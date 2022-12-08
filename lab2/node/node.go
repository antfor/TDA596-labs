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

const m = 4 // Maybe need to change
var maxNodes = new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)

var predecessor *Node = &Node{}

var successor *Node = &Node{}

var next int = 0
var fingers [m]*Node

func (n *Node) Create(_ *Empty, _ *Empty) error {
	predecessor = nil
	successor = n

	// init successor list and finger table appropriately (i.e., all will point to the client itself)
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

	fmt.Println("Find_successor")

	if between(*id, n.Id, successor.Id, true) {

		*nr = *successor
		return nil
	} else {
		n0 := &Node{}
		err := n.Closest_preceding_node(id, n0)

		if err != nil {
			return err
		}

		if *n0 == *n { // ???
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

	err := Call(successor, "Node.GetPredecessor", &Empty{}, x)

	if err != nil {
		return err
	}

	if (*x != Node{}) {

		if between(x.Id, n.Id, successor.Id, false) {
			successor = x
		}
	}

	return Call(successor, "Node.Notify", n, &Empty{})
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

	return Call(n, "Node.Find_successor", &id, fingers[next])
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
	jump := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(finger)-1), nil)
	dest := start.Add(start, jump)

	return new(big.Int).Mod(dest, maxNodes).Text(16)
}

// if inclusive = false: returns true if a < x < b  otherwise false
// if inclusive = true:  returns true if a < x <= b otherwise false
func between(xs string, as string, bs string, inclusive bool) bool {

	x := toBigInt(xs)
	a := toBigInt(as)
	b := toBigInt(bs)

	if a.Cmp(x) == -1 && x.Cmp(b) == -1 {
		return true
	}

	if inclusive && x.Cmp(b) == 0 {
		return true
	}

	return false
}
