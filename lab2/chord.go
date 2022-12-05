package main

import (
	"fmt"
	"lab2/argument"
)

//type Key string

//type IPaddress string
/*
type node struct {
	ip   string
	port int
	key  *big.Int

	keys_to_string map[Key]string

	fingers     []string
	predecessor string
	successors  []string
}*/
/*
func hash(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}
*/
func main2() {
	arg, argType := argument.NewArg()
	fmt.Println(arg, "hej")

	if argType == argument.Create {
		// n := node{}

		// n.createRing()

	} else if argType == argument.Join {

		// n := node{ip: arg.Ja, port: arg.Jp}

		// n.joinRing(n)
	} else {
		panic("invalid arguments")
	}

}

/*
func (n node) createRing() {
	fmt.Println("CREATE")
	n.predecessor = ""
	n.successors = append(n.successors, n.ip)
}

func (n node) joinRing(n2 node) {
	fmt.Println("JOIN")
	n.predecessor = ""
	n.successors = append(n.successors, n2.findSuccessor(n))
}

func lookUp() {

}

func (n node) findSuccessor(n2 node) node {

	return node{}
}

func (n node) closest_preceding_node() {

}

// Stabilze stuff
func stabilze() {

}

func notify() {

}

func fix_fingers() {

}

func check_predessesor() {

}

func storeFile() {

}

func printState() {

}
*/
