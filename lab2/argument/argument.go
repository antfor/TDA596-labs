package argument

import (
	"flag"
	"regexp"
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
	S   bool
	G   bool
}

type ArgType int

// argument types
const (
	NotValid = -1
	Create   = 0
	Join     = 1
)

func GetArg() (Argument, ArgType) {

	arg := setFlags()

	validRange(arg)
	isSpecified(arg)

	validId(arg.I)
	checkPort(arg.P)

	argType := getArgType(arg)

	if argType == Join {
		checkPort(arg.Jp)
	}

	return arg, argType
}

func setFlags() Argument {

	arg := Argument{}

	flag.StringVar(&arg.A, "a", "", "a string")

	flag.IntVar(&arg.P, "p", -1, "an int")

	flag.StringVar(&arg.Ja, "ja", "", "a string")
	flag.IntVar(&arg.Jp, "jp", -1, "an int")

	flag.IntVar(&arg.Ts, "ts", -1, "an int")
	flag.IntVar(&arg.Tff, "tff", -1, "an int")
	flag.IntVar(&arg.Tcp, "tcp", -1, "an int")

	flag.IntVar(&arg.R, "r", -1, "an int")

	flag.StringVar(&arg.I, "i", "", "a string")

	flag.BoolVar(&arg.S, "s", false, "a bool")

	flag.BoolVar(&arg.G, "g", false, "a bool")

	flag.Parse()

	return arg
}

func getArgType(arg Argument) ArgType {

	if arg.Ja == "" && arg.Jp == -1 {
		return Create
	}

	if arg.Ja != "" && arg.Jp != -1 {

		return Join
	}

	panic("Not a valid argument")

	return NotValid
}

func isSpecified(arg Argument) {
	if arg.A == "" {
		panic("IP must be specified")
	}
	if (arg.Ja != "" && arg.Jp == -1) && (arg.Ja == "" && arg.Jp != -1) {
		panic("Ja and Jp must both be specified/unspecified")
	}
}

func validId(id string) {

	if id != "" {
		match, err := regexp.MatchString("([0-9]|[a-f]|[A-F]){1,40}", id)

		if err != nil || !match {
			panic("Invalid identifier")
		}
	}
}

func validRange(arg Argument) {
	checkRange(arg.Ts, 1, 60000)
	checkRange(arg.Tff, 1, 60000)
	checkRange(arg.Tcp, 1, 60000)

	checkRange(arg.R, 1, 32)
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
