package main

import (
	"fmt"
	"os"
)

func main() {
	// load config
	c, err := configLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return
	}
	r, err := runnerNew(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup test: %v\n", err)
		return
	}
	err = r.TestAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Test failed:\n%v\n", err)
	}
	r.Report(os.Stdout)
	// TODO: write junit.xml report
}
