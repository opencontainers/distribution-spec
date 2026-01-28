package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	mainRun(false)
}

func mainRun(legacy bool) {
	// load config
	c, err := configLoad()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return
	}
	c.Legacy = legacy
	// run all tests
	r, err := runnerNew(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup test: %v\n", err)
		return
	}
	_ = r.TestAll()
	// show results
	r.Report(os.Stdout)
	// generate reports
	os.MkdirAll(c.ResultsDir, 0755)
	// write config.yaml
	fh, err := os.Create(filepath.Join(c.ResultsDir, "config.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create config.yaml: %v\n", err)
		return
	}
	_, err = fh.Write([]byte(r.Config.Report()))
	_ = fh.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate config.yaml: %v\n", err)
		return
	}
	// write junit.xml report
	fh, err = os.Create(filepath.Join(c.ResultsDir, "junit.xml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create junit.xml: %v\n", err)
		return
	}
	err = r.ReportJunit(fh)
	_ = fh.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate junit.xml: %v\n", err)
		return
	}
	// write report.html
	fh, err = os.Create(filepath.Join(c.ResultsDir, "report.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create report.html: %v\n", err)
		return
	}
	err = r.ReportHTML(fh)
	_ = fh.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate report.html: %v\n", err)
		return
	}
	if c.Legacy {
		fmt.Fprintf(os.Stderr, "WARNING: \"go test\" is deprecated. Please update to using \"go build\".\n")
	}
}
