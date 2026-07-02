//go:build legacy || !unit_tests

package main

import "testing"

func TestLegacy(t *testing.T) {
	mainRun(true)
}
