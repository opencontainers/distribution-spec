package main

import "errors"

var (
	errAPITestDisabled = errors.New("API is disabled by user configuration")
	errAPITestSkip     = errors.New("API test was skipped")
	errAPITestError    = errors.New("API test encountered an internal error")
	errAPITestFail     = errors.New("API test with a known failure")
	errRegUnsupported  = errors.New("registry does not support the requested API")
)
