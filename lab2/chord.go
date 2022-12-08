package main

import (
	"bufio"
	"fmt"
	"lab2/argument"
	"lab2/node"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"runtime"
	"strconv"
	"time"
)

var me *node.Node

func main() {

	arg, argType := argument.GetArg()

	fmt.Println("args:", arg, "type: ", argType)

	id := node.Hash(arg.A + ":" + strconv.Itoa(arg.P))

	if arg.I != "" {
		id = node.Hash(arg.I)
	}

	me = &node.Node{Ip: arg.A, Port: arg.P, Id: id}
	RPC_server(me)

	if argType == argument.Create {

		err := me.Create(&node.Empty{}, &node.Empty{})

		fmt.Println("create err:", err)

	} else if argType == argument.Join {

		id := node.Hash(arg.Ja + ":" + strconv.Itoa(arg.Jp)) // todo call getId

		Jnode := &node.Node{Ip: arg.Ja, Port: arg.Jp, Id: id}

		err := me.Join(Jnode, &node.Empty{})

		fmt.Println("join err:", err)

	} else {
		panic("invalid arguments")
	}

	go timer(arg.Tcp, func() { me.Check_predessesor(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Ts, func() { me.Stabilze(&node.Empty{}, &node.Empty{}) })

	go timer(arg.Tff, func() { me.Fix_fingers(&node.Empty{}, &node.Empty{}) })

	read_stdin()

}

func read_stdin() {

	reader := bufio.NewReader(os.Stdin)

	for {
		cmd, _ := reader.ReadString('\n')
		fmt.Println(cmd)
		fmt.Println(runtime.NumGoroutine())
	}
}

func timer(timeMs int, f func()) {
	for {
		time.Sleep(time.Duration(timeMs) * time.Millisecond)
		f()
	}
}

func RPC_server(n *node.Node) {

	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(n.Port))

	if err != nil {
		fmt.Println("error in RPC_server: ", err)
	}
	go http.Serve(l, nil)
}
