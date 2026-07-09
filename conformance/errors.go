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

import "errors"

var (
	errAPITestDisabled = errors.New("API is disabled by user configuration")
	errAPITestSkip     = errors.New("API test was skipped")
	errAPITestError    = errors.New("API test encountered an internal error")
	errAPITestFail     = errors.New("API test with a known failure")
	errRegUnsupported  = errors.New("registry does not support the requested API")
)
