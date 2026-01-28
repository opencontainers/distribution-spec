package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type results struct {
	Name     string // name of current runner step, concatenated onto the parent's name
	Children []*results
	Parent   *results
	Status   status
	Errs     []error
	Output   *bytes.Buffer
	Start    time.Time
	Stop     time.Time
	Counts   [statusMax]int
}

func resultsNew(name string, parent *results) *results {
	fullName := name
	if parent != nil && parent.Name != "" {
		fullName = fmt.Sprintf("%s/%s", parent.Name, name)
	}
	return &results{
		Name:   fullName,
		Parent: parent,
		Output: &bytes.Buffer{},
		Start:  time.Now(),
	}
}

func (r *results) Count(s string) int {
	st := statusUnknown
	err := st.UnmarshalText([]byte(s))
	if err != nil || st < 0 || st >= statusMax {
		return -1
	}
	return r.Counts[st]
}

func (r *results) ReportWalkErr(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s%s: %s\n", prefix, r.Name, r.Status)
	if len(r.Children) == 0 && len(r.Errs) > 0 {
		// show errors from leaf nodes
		for _, err := range r.Errs {
			fmt.Fprintf(w, "%s - %s\n", prefix, err.Error())
		}
	}
	if len(r.Children) > 0 {
		for _, child := range r.Children {
			child.ReportWalkErr(w, prefix+"  ")
		}
	}
}

func (r *results) ToJunitTestCases() []junitTest {
	jTests := []junitTest{}
	if len(r.Children) == 0 {
		// return the test case for a leaf node
		jTest := junitTest{
			Name:      r.Name,
			Time:      fmt.Sprintf("%f", r.Stop.Sub(r.Start).Seconds()),
			SystemErr: r.Output.String(),
			Status:    r.Status.ToJunit(),
		}
		if len(r.Errs) > 0 {
			jTest.SystemOut = fmt.Sprintf("%v", errors.Join(r.Errs...))
		}
		jTests = append(jTests, jTest)
	}
	if len(r.Children) > 0 {
		// recursively collect test cases from child nodes
		for _, child := range r.Children {
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
		return "FAIL"
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

func (s *status) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "pass":
		*s = statusPass
	case "skip":
		*s = statusSkip
	case "disabled":
		*s = statusDisabled
	case "fail":
		*s = statusDisabled
	case "error":
		*s = statusError
	case "unknown":
		*s = statusUnknown
	default:
		return fmt.Errorf("unknown status %s", string(text))
	}
	return nil
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
