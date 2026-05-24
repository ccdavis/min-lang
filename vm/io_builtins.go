package vm

import (
	"bufio"
	"fmt"
	"os"
)

var stdinScanner *bufio.Scanner

func init() {
	stdinScanner = bufio.NewScanner(os.Stdin)
}

func writeBuiltin(args ...Value) Value {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.String())
	}
	return NilValue()
}

func eprintBuiltin(args ...Value) Value {
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg.String())
	}
	fmt.Fprintln(os.Stderr)
	return NilValue()
}

func ewriteBuiltin(args ...Value) Value {
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg.String())
	}
	return NilValue()
}

func readlnBuiltin(args ...Value) Value {
	if len(args) > 0 {
		fmt.Print(args[0].String())
	}
	if stdinScanner.Scan() {
		return StringValue(stdinScanner.Text())
	}
	if err := stdinScanner.Err(); err != nil {
		printError("readln: %s", err.Error())
	}
	return NilValue()
}
