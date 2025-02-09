package main

import "encoding/xml"

const (
	junitPassed  = "passed"  // successful test
	junitSkipped = "skipped" // test intentionally skipped
	junitFailure = "failure" // test ran but failed, e.g. missed assertion
	junitError   = "error"   // test encountered an unexpected error
)

type junitProperty struct {
	Name  string `xml:"name,attr"`  // name or key
	Value string `xml:"value,attr"` // value of name
}

type junitResult struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr,omitempty"`
	Data    string `xml:",cdata"`
}

type junitTest struct {
	Name      string       `xml:"name,attr"`             // name of the test
	Classname string       `xml:"classname,attr"`        // hierarch of test
	Time      string       `xml:"time,attr,omitempty"`   // duration in seconds
	Status    string       `xml:"status,attr,omitempty"` // passed, skipped, failure, or error
	Skipped   *junitResult `xml:"skipped,omitempty"`     // result from skipped tests
	Failure   *junitResult `xml:"failure,omitempty"`     // result from test failures
	Error     *junitResult `xml:"error,omitempty"`       // result from test errors
	SystemOut string       `xml:"system-out,omitempty"`  // output written to stdout
	SystemErr string       `xml:"system-err,omitempty"`  // output written to stderr
}

type junitTestSuite struct {
	Name       string          `xml:"name,attr"`                     // name of suite
	Package    string          `xml:"package,attr,omitempty"`        // hierarchy of suite
	Tests      int             `xml:"tests,attr"`                    // count of tests
	Failures   int             `xml:"failures,attr"`                 // count of failures
	Errors     int             `xml:"errors,attr"`                   // count of errors
	Disabled   int             `xml:"disabled,attr,omitempty"`       // count of disabled tests
	Skipped    int             `xml:"skipped,attr,omitempty"`        // count of skipped tests
	Time       string          `xml:"time,attr"`                     // duration in seconds
	Timestamp  string          `xml:"timestamp,attr,omitempty"`      // ISO8601
	Properties []junitProperty `xml:"properties>property,omitempty"` // mapping of key/value pairs associated with the test
	Testcases  []junitTest     `xml:"testcase,omitempty"`            // slice of tests
	SystemOut  string          `xml:"system-out,omitempty"`          // output written to stdout
	SystemErr  string          `xml:"system-err,omitempty"`          // output written to stderr
}

type junitTestSuites struct {
	XMLName  xml.Name         `xml:"testsuites"`              // xml namespace and name
	Name     string           `xml:"name,attr,omitempty"`     // name of the collection of suites
	Time     string           `xml:"time,attr,omitempty"`     // duration in seconds
	Tests    int              `xml:"tests,attr,omitempty"`    // count of tests
	Errors   int              `xml:"errors,attr,omitempty"`   // count of errors
	Failures int              `xml:"failures,attr,omitempty"` // count of failures
	Skipped  int              `xml:"skipped,attr,omitempty"`  // count of skipped tests
	Disabled int              `xml:"disabled,attr,omitempty"` // count of disabled tests
	Suites   []junitTestSuite `xml:"testsuite,omitempty"`     // slice of suites
}
