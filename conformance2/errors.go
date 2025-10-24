package main

import "errors"

var (
	ErrDisabled       = errors.New("test is disabled")
	ErrRegUnsupported = errors.New("registry does not support the requested API")
)
