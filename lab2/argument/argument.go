package argument

import (
	"fmt"
	"os"
	"strconv"
)

type Argument struct {
	A   string
	P   int
	Ja  string
	Jp  int
	Ts  int
	Tff int
	Tcp int
	R   int
	I   string
}

const (
	Create = 0
	Join   = 1
)

func NewArg() (Argument, int) {

	arg := Argument{}

	for i := 0; i < len(os.Args); i += 2 {
		fmt.Println(i, os.Args[i])

		switch os.Args[i] {
		case "-a":
			arg.A = os.Args[i+1]
		case "-p":
			arg.P = getPort(os.Args[i+1])
		case "--ja":
			arg.Ja = os.Args[i+1]
		case "--jp":
			arg.Jp = getPort(os.Args[i+1])
		case "--ts":
			arg.Ts, _ = strconv.Atoi(os.Args[i+1]) // todo fix 1-60000
		case "--tff":
			arg.Tff, _ = strconv.Atoi(os.Args[i+1]) // todo fix 1-60000
		case "--tcp":
			arg.Tcp, _ = strconv.Atoi(os.Args[i+1]) // todo fix 1-60000
		case "-r":
			arg.R, _ = strconv.Atoi(os.Args[i+1]) // todo fix 1-32
		case "-i":
			arg.I = os.Args[i+1] // todo fix string of 40 characters matching [0-9a-fA-F]

		}
	}

	//validArgs() TODO

	if arg.Ja == "" && arg.Jp == 0 {
		return arg, Create
	} else if arg.Ja != "" && arg.Jp != 0 {
		return arg, Join
	}

	return arg, -1
}

func getPort(sPort string) int {

	portNum, err := strconv.Atoi(sPort)

	if err != nil || portNum < 0 || portNum > 65535 {
		panic("Error in argument: not a valid port")
	}

	return portNum
}
