package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// load config
	c, err := configLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return
	}
	// run all tests
	r, err := runnerNew(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup test: %v\n", err)
		return
	}
	_ = r.TestAll()
	// show results
	r.Report(os.Stdout)
	os.MkdirAll(c.ResultsDir, 0755)
	// write junit.xml report
	ju := r.ToJunit()
	fh, err := os.Create(filepath.Join(c.ResultsDir, "junit.xml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create junit.xml: %v\n", err)
		return
	}
	enc := xml.NewEncoder(fh)
	enc.Indent("", "  ")
	err = enc.Encode(ju)
	_ = fh.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate junit.xml: %v\n", err)
		return
	}
	// TODO: write report.html
}
