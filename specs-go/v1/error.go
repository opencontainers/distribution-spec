// Copyright 2019 The Linux Foundation
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

package v1

// ErrorResponse is returned by a registry on an invalid request.
type ErrorResponse struct {
	Errors []ErrorInfo `json:"errors"`
}

// ErrRegistry is the string returned by and ErrorResponse error.
var ErrRegistry = "distribution: registry returned error"

// Error implements the Error interface.
func (er *ErrorResponse) Error() string {
	return ErrRegistry
}

// Detail returns an ErrorInfo
func (er *ErrorResponse) Detail() []ErrorInfo {
	return er.Errors
}

// ErrorInfo describes a server error returned from a registry.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}
