package node

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math/big"
	"net/rpc"
	"sort"
	"strconv"
)

type Node struct {
	Ip string

	Port int

	Id string

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
var conn io.Writer

func (n *Node) SetConn(c io.Writer) {
	fmt.Println("SetConn", c)
	conn = c
}

var predecessor *Node = &Node{}

var next int = 0
var fingers [m]*Node

var successors []*Node

//var FileMap map[string]string // Map a key to a file name

type FileAndBackups struct {
	File   string
	Backup []*Node
	Key    string
}

var FileMap map[string]FileAndBackups // Map a key to a file name

func (n *Node) GetSuccessors() []*Node {
	return successors
}

func (n *Node) init() {

	FileMap = make(map[string]FileAndBackups)

	successors = make([]*Node, n.R)

	for i := 0; i < n.R; i++ {
		successors[i] = n
	}

	for i := 0; i < m; i++ {
		fingers[i] = n
	}

}

func (n *Node) Create(r *int, _ *Empty) error {
	n.init()

	predecessor = nil
	setSuccessor(n)

	return nil
}

func (n *Node) Join(np *Node, _ *Empty) error {
	n.init()

	predecessor = nil

	reply := &Node{Ip: "start"}
	err := Call(np, "Node.Find_successor", &n.Id, reply)

	if err != nil {
		return err
	}

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
			fmt.Fprintln(conn, "error in find_suc1: ", err)
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

// Notify the node that it thinks it is your predecessor
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

// Fixes the finger table entries
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

// Checks if the predecessor is alive
// If not, sets the predecessor to nil
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

// Returns "pong" if the node is running
// Used to check if a node is alive
func (n *Node) Ping(_ *Empty, reply *string) error {

	*reply = "pong"
	return nil
}

// TODO ADD COMMENTS!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (n *Node) Stabilze(_ *Empty, _ *Empty) error {

	xd := PredAndSuccList{Pred: Node{}, Succ: []*Node{}}

	err := Call(fingers[0], "Node.GetPredAndSuccessors", &Empty{}, &xd)

	if err != nil {
		fmt.Fprintln(conn, "error in stab, deleting node:", err)

		if len(successors) > 1 {
			removeAllBackups(n, successors[0], successors[1])
		}

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

// Helper functions /////////////////////////////////////////////////

// Replace your successor with the new successor
// But keep set your successor first
// Shorten it to R, if needed
func (n *Node) replaceSuccessors(Succ []*Node) {

	tempList := []*Node{fingers[0]}
	tempList = append(tempList, Succ...)

	if !(len(tempList) < n.R) {
		successors = tempList[:n.R]
	} else {
		successors = tempList
	}

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

// Get the predecessor and successors of the node
func (n *Node) GetPredAndSuccessors(_ *Empty, nd *PredAndSuccList) error {

	if predecessor != nil {
		nd.Pred = *predecessor
	}

	nd.Succ = successors

	return nil
}

// Set the successor of the node
func setSuccessor(s *Node) {

	tmp := append([]*Node{s}, successors...)

	if len(successors) >= s.R {
		tmp = tmp[:s.R]
	}
	successors = tmp
	fingers[0] = s

}

// Returns the node that is running this code
func (n *Node) GetMe(_ *Empty, node *Node) error {

	*node = *n
	return nil
}

// Returns the node at the given ip and port
func GetNode(ip string, port int, node *Node) error {

	err := Call(&Node{Ip: ip, Port: port}, "Node.GetMe", &Empty{}, node)

	return err
}

// Does a RPC call to node n, with function f, with argument arg, and reply reply
func Call[A any, R any](n *Node, f string, arg *A, reply *R) error {

	client, err := rpc.DialHTTP("tcp", n.Ip+":"+strconv.Itoa(n.Port))

	if err != nil {
		return err
	}

	defer client.Close()
	return client.Call(f, arg, reply)
}

// todo remove
func (n *Node) StoreFileAtNode(node *Node, filename File) error {
	err := Call(node, "Node.StoreFile", &filename, &Empty{})
	if err != nil {
		fmt.Fprintln(conn, "Error in StoreFileAtNode: ", err)
		return err
	}
	return nil
}

// Returns the hex value of a hash of a string, mod 2^m
func Hash(elt string) string {
	hasher := sha1.New()
	hasher.Write([]byte(elt))

	hashValue := new(big.Int).SetBytes(hasher.Sum(nil))

	key := new(big.Int).Mod(hashValue, maxNodes)

	return key.Text(16)
}

// Mod(id) = id % 2^m
func Mod(id string) string {
	keyID := new(big.Int)
	keyID.SetString(id, 16)
	keyID.Mod(keyID, maxNodes)

	return keyID.Text(16)
}

// Turns a string into a big int
func toBigInt(key string) *big.Int {
	keyID := new(big.Int)
	keyID.SetString(key, 16)

	return keyID
}

// Jump(startID, finger) = (startID + 2^finger) % 2^m
func jump(startID string, finger int) string {
	start := toBigInt(startID)
	jump := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(finger)), nil)
	dest := start.Add(start, jump)

	return new(big.Int).Mod(dest, maxNodes).Text(16)
}

// if inclusive = false: returns true if a < x < b  otherwise false
// if inclusive = true:  returns true if a < x <= b otherwise false
// if inclusive = false and a > b: returns true if a < x or x < b  otherwise false
// if inclusive = true  and a > b: returns true if a < x or x <= b otherwise false
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

// CMD /////////////////////////////////////////////

type File struct {
	Key  string
	File string
}

func (n *Node) StoreFile(fileAndKey *File, _ *Empty) error {
	key := fileAndKey.Key
	file := fileAndKey.File

	n.Print()
	fmt.Fprintln(conn, "key: ", key)
	fmt.Fprintln(conn, "file: ", file)
	fmt.Fprintln(conn, "store: ", FileMap)

	backup := make([]*Node, 1)
	backup[0] = n

	// todo add successors to backup
	for i := 0; i < len(successors)-1; i++ {
		backup = append(backup, successors[i])
	}
	fb := FileAndBackups{File: file, Key: key, Backup: backup}

	for i := 0; i < len(successors)-1; i++ {
		Call(successors[i], "Node.NewBackup", &fb, &Empty{})
	}

	FileMap[key] = fb

	return nil
}

func (n *Node) TakeFiles() (*Node, []string) {

	s := fingers[0]

	fb := make([]FileAndBackups, 0)

	Call(s, "Node.TakeKeys", n, &fb)

	files := make([]string, 0)

	for _, file := range fb {
		files = append(files, file.File)
		fixBackup(n, file)
	}

	return s, files
}

type NodeList []*Node

func (ns NodeList) Len() int {
	return len(ns)
}

func (ns NodeList) Less(i, j int) bool {
	iId := toBigInt(ns[i].Id)
	jId := toBigInt(ns[j].Id)

	return iId.Cmp(jId) == 1
}

func (ns NodeList) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func remove(s []*Node, e *Node) ([]*Node, bool) {
	for i, a := range s {
		if *a == *e {
			return append(s[:i], s[i+1:]...), true
		}
	}
	return s, false
}

func removeAllBackups(me *Node, succ *Node, succSucc *Node) {
	err := me.RemoveBackups(succ, &Empty{})
	if err != nil {
		fmt.Fprintln(conn, "error in removeAllBackups1: ", err)
	}

	err = Call(succSucc, "Node.RemoveBackups", succ, &Empty{})

	if err != nil {
		fmt.Fprintln(conn, "error in removeAllBackups2: ", err)
	}
}

func (n *Node) RemoveBackups(succ *Node, _ *Empty) error {

	for _, fb := range FileMap {
		var removed bool
		fb.Backup, removed = remove(fb.Backup, succ)

		if removed {
			for _, node := range fb.Backup {
				err := Call(node, "Node.NewBackup", &fb, &Empty{})
				if err != nil {
					fmt.Fprintln(conn, "error in RemoveBackups: ", err)
				}
			}
		}
	}

	return nil
}

func fixBackup(me *Node, fb FileAndBackups) {

	fmt.Fprintln(conn, "fixBackup: ")

	backups := fb.Backup
	key := fb.Key

	backups = append(backups, me)
	fmt.Fprintln(conn, "backups: ", backups)

	sort.Sort(NodeList(backups))

	var toDelete []*Node

	if len(backups) > me.R {
		toDelete = backups[me.R-1:]
		backups = backups[:me.R]
	}

	newFb := FileAndBackups{Key: key, File: fb.File, Backup: backups}

	for _, node := range toDelete {
		Call(node, "Node.DeleteKey", &key, &Empty{})
	}

	for _, node := range backups {
		Call(node, "Node.NewBackup", &newFb, &Empty{})
	}
	//return toDelete
}

func (n *Node) DeleteKey(key *string, _ *Empty) error {
	delete(FileMap, *key)
	return nil
}

func (n *Node) NewBackup(newBackups *FileAndBackups, _ *Empty) error {
	FileMap[newBackups.Key] = *newBackups
	return nil
}

// key < me < succ

func (n *Node) TakeKeys(me *Node, fb *[]FileAndBackups) error {

	for key, file := range FileMap {
		if between(me.Id, key, n.Id, false) {

			fixBackup(me, file)
			//fmt.Fprintln(conn, "Delkey: ", key)

			//*fb = append(*fb, file)
		}
	}

	return nil
}

func (n *Node) GetFileMap(_ *Empty, filemap *map[string]FileAndBackups) error {
	*filemap = FileMap
	return nil
}

// CMD PrintState
func (n *Node) PrintState() {
	fmt.Fprintln(conn, "me: ")
	n.Print()
	if predecessor != nil {
		fmt.Fprintln(conn, "predecessor: ")
		predecessor.Print()
	} else {
		fmt.Fprintln(conn, "predecessor is nil ")
	}

	fmt.Fprintln(conn, "successor: ")

	fingers[0].Print()
	fmt.Fprintln(conn, "finger table: ")

	for i := 0; i < m; i++ {
		if fingers[i] != nil {
			fmt.Fprintln(conn, "finger", i, ":")
			fingers[i].Print()
		} else {
			fmt.Fprintln(conn, "finger ", i, " is nil")
		}
	}

	fmt.Fprintln(conn, "Successors: ", successors)
	fmt.Print("    ")
	for _, s := range successors {
		fmt.Print(s.Id, " , ")
	}
	fmt.Fprintln(conn)

	fmt.Fprintln(conn, "Files: ", FileMap)

	for key, file := range FileMap {
		fmt.Fprintln(conn, "	key: ", key, " file: ", file.File)
		for _, b := range file.Backup {
			fmt.Print("		backup: ", b.Id)
		}
	}
}

// Prints the node information
func (n *Node) Print() {
	fmt.Fprintln(conn, "	ip: ", n.Ip)
	fmt.Fprintln(conn, "	port: ", n.Port)

	value, _ := strconv.ParseInt(n.Id, 16, 64)
	fmt.Fprintln(conn, "	id (base 10)", value)
}
