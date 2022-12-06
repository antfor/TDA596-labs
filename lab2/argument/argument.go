package argument

import (
	"flag"
	"regexp"
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

	aPtr := flag.String("a", "", "a string")

	pPtr := flag.Int("p", -1, "an int")

	jaPtr := flag.String("ja", "", "a string")
	jpPtr := flag.Int("jp", -1, "an int")

	tsPtr := flag.Int("ts", -1, "an int")
	tffPtr := flag.Int("tff", -1, "an int")
	tcpPtr := flag.Int("tcp", -1, "an int")

	rPtr := flag.Int("r", -1, "an int")

	iPtr := flag.String("i", "", "a string")

	flag.Parse()

	arg.A = *aPtr

	arg.P = checkPort(*pPtr)

	arg.Ja = *jaPtr
	arg.Jp = *jpPtr

	arg.Ts = checkRange(*tsPtr, 1, 60000)
	arg.Tff = checkRange(*tffPtr, 1, 60000)
	arg.Tcp = checkRange(*tcpPtr, 1, 60000)

	arg.R = checkRange(*rPtr, 1, 32)

	if *iPtr != "" {
		match, err := regexp.MatchString("([0-9]|[a-f]|[A-F]){40}", *iPtr)

		if err != nil || !match {
			panic("Invalid identifier")
		}
		arg.I = *iPtr
	}

	if arg.A == "" {
		panic("IP must be specified")
	}
	if (arg.Ja != "" && arg.Jp == -1) && (arg.Ja == "" && arg.Jp != -1) {
		panic("Invalid Argument")
	}

	if arg.Ja == "" && arg.Jp == -1 {
		return arg, Create
	}

	arg.Jp = checkPort(arg.Jp)

	if arg.Ja != "" && arg.Jp != -1 {

		return arg, Join
	}

	return arg, -1
}

func checkRange(time int, min int, max int) int {
	if min <= time && time <= max {
		return time
	}

	panic("Invalid Argument")
}

func checkPort(portNum int) int {

	if portNum < 0 || portNum > 65535 {
		panic("Error in argument: not a valid port")
	}

	return portNum
}

func getPort(sPort string) int {

	portNum, err := strconv.Atoi(sPort)

	if err != nil || portNum < 0 || portNum > 65535 {
		panic("Error in argument: not a valid port")
	}

	return portNum
}
