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

import (
	"fmt"
)

// InvalidInput creates an error for invalid input parameters
func InvalidInput(reason string, args ...interface{}) *ErrorInfo {
	return &ErrorInfo{
		Code:    InvalidInputCode,
		Message: invalidInputMsg,
		Reason:  fmt.Sprintf(reason, args...),
	}
}

// ServerError creates an error for internal server errors
func ServerError(reason string, args ...interface{}) *ErrorInfo {
	return &ErrorInfo{
		Code:    ServerErrorCode,
		Message: serverErrorMsg,
		Reason:  fmt.Sprintf(reason, args...),
	}
}
