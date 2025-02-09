package main

import (
	"bytes"
	"fmt"
	"time"
)

type results struct {
	status status
	errs   []error
	output *bytes.Buffer
	start  time.Time
	stop   time.Time
	counts [statusMax]int
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
