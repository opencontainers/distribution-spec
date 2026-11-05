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

//go:build unit_tests

package main

import "testing"

func TestParseAuthHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		expect authHeader
	}{
		{
			name:   "basic",
			header: `Basic realm="registry.example.com"`,
			expect: authHeader{Type: "basic"},
		},
		{
			name:   "bearer",
			header: `Bearer realm="https://auth.example.com/token",service="registry.example.com",scope="repository:conformance:pull"`,
			expect: authHeader{Type: "bearer", Realm: "https://auth.example.com/token", Service: "registry.example.com", Scope: "repository:conformance:pull"},
		},
		{
			name:   "bearer lowercase",
			header: `bearer realm="https://auth.example.com/token",service="registry.example.com"`,
			expect: authHeader{Type: "bearer", Realm: "https://auth.example.com/token", Service: "registry.example.com"},
		},
		{
			name:   "bearer with basic in scope",
			header: `Bearer realm="https://auth.example.com/token",service="registry.example.com",scope="repository:conformance/basic:pull"`,
			expect: authHeader{Type: "bearer", Realm: "https://auth.example.com/token", Service: "registry.example.com", Scope: "repository:conformance/basic:pull"},
		},
		{
			name:   "bearer with basic in realm",
			header: `Bearer realm="https://basic.example.com/token",service="registry.example.com"`,
			expect: authHeader{Type: "bearer", Realm: "https://basic.example.com/token", Service: "registry.example.com"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parseAuthHeader(tc.header)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed != tc.expect {
				t.Errorf("unexpected result, expected %+v, received %+v", tc.expect, parsed)
			}
		})
	}
}
