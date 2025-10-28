// Copyright The AIGW Authors.
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

package errors

import "fmt"

// ErrorInfo represents a structured error response
// Used for consistent error handling across the application
type ErrorInfo struct {
	Code    int    `json:"code"`    // Error code for programmatic handling
	Message string `json:"message"` // User-friendly error message
	Reason  string `json:"reason"`  // Technical reason for debugging
}

// Error implements the error interface
// Returns a formatted string representation of the error
func (r *ErrorInfo) Error() string {
	return fmt.Sprintf(`errCode=%d, errMsg=%s, errReason=%s`, r.Code, r.Message, r.Reason)
}

// GetStatusCode extracts HTTP status code from error code
// Returns the first 3 digits of the error code as HTTP status
func (r *ErrorInfo) GetStatusCode() int {
	return r.Code / 100000
}

// SetMassage sets the user-friendly error message
// Returns the ErrorInfo for method chaining
func (r *ErrorInfo) SetMassage(msg string) *ErrorInfo {
	r.Message = msg
	return r
}
