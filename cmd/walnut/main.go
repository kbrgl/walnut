package main

import (
	"fmt"
	"os"

	"github.com/kbrgl/walnut"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: walnut [file]")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	c := walnut.NewCompiler(os.Stdout)
	if err = c.Compile(f, 30000, walnut.PtrCenter); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
