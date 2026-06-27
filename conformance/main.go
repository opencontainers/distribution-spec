// Copyright the Open Container Initiative Contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
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
	err = r.TestAll()
	if err != nil && !errors.Is(err, errRegUnsupported) && !errors.Is(err, errAPITestFail) &&
		!errors.Is(err, errAPITestSkip) && !errors.Is(err, errAPITestDisabled) {
		fmt.Fprintf(os.Stderr, "failed to run tests: %v", err)
	}
	// show results
	r.Report(os.Stdout)
	// generate reports
	err = os.MkdirAll(c.ResultsDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed create results directory: %v\n", err)
		return
	}
	// write results.yaml
	fh, err := os.Create(filepath.Join(c.ResultsDir, "results.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create results.yaml: %v\n", err)
		return
	}
	err = r.ReportResultsYAML(fh)
	_ = fh.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate results.yaml: %v\n", err)
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
	if r.Results.Status != statusPass {
		fmt.Fprintf(os.Stderr, "*** Conformance test detected a failure. ***\n")
		os.Exit(1)
	}
}
