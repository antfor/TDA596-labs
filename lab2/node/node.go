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

	FileMap map[string]string // Map a key to a file name

	R int
}

type Empty struct {
}

type PredAndSuccList struct {
	Pred Node
	Succ []*Node
}

const m = 2 // Maybe need to change
var maxNodes = new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)

var predecessor *Node = &Node{}

var next int = 0
var fingers [m]*Node

var successors []*Node

//var fileMap map[string]string // Map a key to a file name

func (n *Node) init() {

	n.FileMap = make(map[string]string)

	successors = make([]*Node, 0)

	//todo: init finger table and successor list
}

func (n *Node) StoreFile(key string, file string) {
	n.Print()
	fmt.Println("key: ", key)
	fmt.Println("file: ", file)
	fmt.Println("store: ", n.FileMap)

	n.FileMap[key] = file
}

func (n *Node) TakesKeys() (*Node, []string) {
	s := fingers[0]
	var files []string

	for key, file := range s.FileMap {
		if between(n.Id, key, s.Id, false) {
			n.FileMap[key] = file
			delete(s.FileMap, key)
			files = append(files, file)
		}
	}
	return s, files
}

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

	fmt.Println("Successors: ", successors)
	fmt.Print("    ")
	for _, s := range successors {
		fmt.Print(s.Id, " , ")
	}
	fmt.Println()

}

func (n *Node) Create(r *int, _ *Empty) error {
	predecessor = nil

	//fingers[0] = n

	n.init()

	setSuccessor(n)

	return nil
}

func (n *Node) Join(np *Node, _ *Empty) error {

	predecessor = nil
	n.init()

	reply := &Node{Ip: "start"}
	err := Call(np, "Node.Find_successor", &n.Id, reply)

	if err != nil {
		return err
	}

	setSuccessor(n)
	setSuccessor(reply)
	return nil
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

		if n0.Id == n.Id {
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

func (n *Node) GetPredAndSuccessors(_ *Empty, nd *PredAndSuccList) error {

	if predecessor != nil {
		nd.Pred = *predecessor
	}

	nd.Succ = successors

	return nil
}

func (n *Node) replaceSuccessors(Succ []*Node) {

	tempList := []*Node{fingers[0]}
	tempList = append(tempList, Succ...)

	if !(len(tempList) < n.R) {
		successors = tempList[:n.R]
	} else {
		successors = tempList
	}

}

func (n *Node) Stabilze(_ *Empty, _ *Empty) error {

	xd := PredAndSuccList{Pred: Node{}, Succ: []*Node{}}

	err := Call(fingers[0], "Node.GetPredAndSuccessors", &Empty{}, &xd)

	if err != nil {
		fmt.Println("error in stab, deleting node:", err)

		newSuccessor(n)
		err = n.Stabilze(&Empty{}, &Empty{})

		return err
	}

	n.replaceSuccessors(xd.Succ)
	x := &xd.Pred

	if (x.Id != Node{}.Id) {

		if between(x.Id, n.Id, fingers[0].Id, false) {
			setSuccessor(x)
		}
	}

	return Call(fingers[0], "Node.Notify", n, &Empty{})
}

// Chop the first element off your successors list,
// and set your successor to the next element in the list.
// If there is no such element (the list is empty), set your successor to your own address.
func newSuccessor(me *Node) {

	if len(successors) > 1 {
		successors = successors[1:]
		fingers[0] = successors[0]
	} else if len(successors) == 1 {
		successors = successors[1:]
		setSuccessor(me)
	} else {
		setSuccessor(me)
	}

}

func setSuccessor(s *Node) {

	successors = append([]*Node{s}, successors...)
	fmt.Println("setting successor to ", successors)
	fingers[0] = s

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

func (n *Node) GetMe(_ *Empty, node *Node) error {

	*node = *n
	return nil
}

func GetNode(ip string, port int, node *Node) error {

	err := Call(&Node{Ip: ip, Port: port}, "Node.GetMe", &Empty{}, node)

	return err
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

func Mod(id string) string {
	keyID := new(big.Int)
	keyID.SetString(id, 16)
	keyID.Mod(keyID, maxNodes)

	return keyID.Text(16)
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
