package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

type results struct {
	name     string // name of current runner step, concatenated onto the parent's name
	children []*results
	parent   *results
	status   status
	errs     []error
	output   *bytes.Buffer
	start    time.Time
	stop     time.Time
	counts   [statusMax]int
}

func resultsNew(name string, parent *results) *results {
	fullName := name
	if parent != nil && parent.name != "" {
		fullName = fmt.Sprintf("%s/%s", parent.name, name)
	}
	return &results{
		name:   fullName,
		parent: parent,
		output: &bytes.Buffer{},
		start:  time.Now(),
	}
}

func (r *results) Start() {
	r.start = time.Now()
}

func (r *results) ReportWalkErr(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s%s: %s\n", prefix, r.name, r.status)
	if len(r.children) == 0 && len(r.errs) > 0 {
		// show errors from leaf nodes
		for _, err := range r.errs {
			fmt.Fprintf(w, "%s - %s\n", prefix, err.Error())
		}
	}
	if len(r.children) > 0 {
		for _, child := range r.children {
			child.ReportWalkErr(w, prefix+"  ")
		}
	}
}

func (r *results) ToJunit() *junitTestSuites {
	statusTotal := 0
	for i := status(1); i < statusMax; i++ {
		statusTotal += r.counts[i]
	}
	tSec := fmt.Sprintf("%f", r.stop.Sub(r.start).Seconds())
	jTSuites := junitTestSuites{
		Tests:    statusTotal,
		Errors:   r.counts[statusError],
		Failures: r.counts[statusFail],
		Skipped:  r.counts[statusSkip],
		Disabled: r.counts[statusDisabled],
		Time:     tSec,
	}
	jTSuite := junitTestSuite{
		Name:     r.name,
		Tests:    statusTotal,
		Errors:   r.counts[statusError],
		Failures: r.counts[statusFail],
		Skipped:  r.counts[statusSkip],
		Disabled: r.counts[statusDisabled],
		Time:     tSec,
	}
	jTSuite.Testcases = r.ToJunitTestCases()
	// TODO: inject configuration as properties on jTSuite
	jTSuites.Suites = []junitTestSuite{jTSuite}
	return &jTSuites
}

func (r *results) ToJunitTestCases() []junitTest {
	jTests := []junitTest{}
	if len(r.children) == 0 {
		// return the test case for a leaf node
		jTest := junitTest{
			Name:      r.name,
			Time:      fmt.Sprintf("%f", r.stop.Sub(r.start).Seconds()),
			SystemErr: r.output.String(),
			Status:    r.status.ToJunit(),
		}
		if len(r.errs) > 0 {
			jTest.SystemOut = fmt.Sprintf("%v", errors.Join(r.errs...))
		}
		jTests = append(jTests, jTest)
	}
	if len(r.children) > 0 {
		// recursively collect test cases from child nodes
		for _, child := range r.children {
			jTests = append(jTests, child.ToJunitTestCases()...)
		}
	}
	return jTests
}

type status int

const (
	statusUnknown  status = iota // status is undefined
	statusDisabled               // test was disabled by configuration
	statusSkip                   // test was skipped
	statusPass                   // test passed
	statusFail                   // test detected a conformance failure
	statusError                  // failure of the test engine itself
	statusMax                    // only used for allocating arrays
)

func (s status) Set(set status) status {
	// only set status to a higher level
	if set > s {
		return set
	}
	return s
}

func (s status) String() string {
	switch s {
	case statusPass:
		return "Pass"
	case statusSkip:
		return "Skip"
	case statusDisabled:
		return "Disabled"
	case statusFail:
		return "Fail"
	case statusError:
		return "Error"
	default:
		return "Unknown"
	}
}

func (s status) MarshalText() ([]byte, error) {
	ret := s.String()
	if ret == "Unknown" {
		return []byte(ret), fmt.Errorf("unknown status %d", s)
	}
	return []byte(ret), nil
}

func (s status) ToJunit() string {
	switch s {
	case statusPass:
		return junitPassed
	case statusSkip, statusDisabled:
		return junitSkipped
	case statusFail:
		return junitFailure
	default:
		return junitError
	}
}
