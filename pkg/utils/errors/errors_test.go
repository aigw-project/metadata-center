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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		args   []interface{}
		expect *ErrorInfo
	}{
		{
			name:   "simple reason",
			reason: "invalid parameter",
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: invalidInputMsg,
				Reason:  "invalid parameter",
			},
		},
		{
			name:   "reason with formatting",
			reason: "invalid parameter: %s",
			args:   []interface{}{"test"},
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: invalidInputMsg,
				Reason:  "invalid parameter: test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InvalidInput(tt.reason, tt.args...)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestServerError(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		args   []interface{}
		expect *ErrorInfo
	}{
		{
			name:   "simple reason",
			reason: "database connection failed",
			expect: &ErrorInfo{
				Code:    ServerErrorCode,
				Message: serverErrorMsg,
				Reason:  "database connection failed",
			},
		},
		{
			name:   "reason with formatting",
			reason: "failed to connect to %s",
			args:   []interface{}{"database"},
			expect: &ErrorInfo{
				Code:    ServerErrorCode,
				Message: serverErrorMsg,
				Reason:  "failed to connect to database",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ServerError(tt.reason, tt.args...)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestErrorInfo_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *ErrorInfo
		expect   string
	}{
		{
			name: "basic error info",
			error: &ErrorInfo{
				Code:    40001400,
				Message: "Invalid input parameters",
				Reason:  "invalid parameter",
			},
			expect: "errCode=40001400, errMsg=Invalid input parameters, errReason=invalid parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestErrorInfo_GetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		error    *ErrorInfo
		expect   int
	}{
		{
			name: "invalid input code",
			error: &ErrorInfo{
				Code: InvalidInputCode,
			},
			expect: 400,
		},
		{
			name: "server error code",
			error: &ErrorInfo{
				Code: ServerErrorCode,
			},
			expect: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.GetStatusCode()
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestErrorInfo_SetMassage(t *testing.T) {
	tests := []struct {
		name     string
		error    *ErrorInfo
		message  string
		expect   *ErrorInfo
	}{
		{
			name: "set message",
			error: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "original message",
				Reason:  "test reason",
			},
			message: "new message",
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "new message",
				Reason:  "test reason",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.SetMassage(tt.message)
			assert.Equal(t, tt.expect, result)
			assert.Equal(t, tt.expect, tt.error) // Should modify in place
		})
	}
}

func TestParseJSONFailed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expect   *ErrorInfo
	}{
		{
			name: "JSON syntax error",
			err:  errors.New("unexpected EOF, looking for beginning of"),
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "Invalid input parameters: invalid JSON format",
				Reason:  "unexpected EOF, looking for beginning of",
			},
		},
		{
			name: "type mismatch error",
			err:  errors.New("cannot unmarshal string into Go struct field User.age of type int"),
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "Invalid input parameters: field User.age should be int, not string",
				Reason:  "cannot unmarshal string into Go struct field User.age of type int",
			},
		},
		{
			name: "validation error - max",
			err:  errors.New("AddError:Field validation for 'Age' failed on the 'max' tag"),
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "Invalid input parameters: field Age exceeds maximum value",
				Reason:  "AddError:Field validation for 'Age' failed on the 'max' tag",
			},
		},
		{
			name: "validation error - required",
			err:  errors.New("AddError:Field validation for 'Name' failed on the 'required' tag"),
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "Invalid input parameters: field Name is required",
				Reason:  "AddError:Field validation for 'Name' failed on the 'required' tag",
			},
		},
		{
			name: "unknown validation error",
			err:  errors.New("AddError:Field validation for 'Field' failed on the 'unknown' tag"),
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "Invalid input parameters: field Field failed validation: unknown",
				Reason:  "AddError:Field validation for 'Field' failed on the 'unknown' tag",
			},
		},
		{
			name: "generic error",
			err:  errors.New("some generic error"),
			expect: &ErrorInfo{
				Code:    InvalidInputCode,
				Message: "Invalid input parameters",
				Reason:  "some generic error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseJSONFailed(tt.err)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestExtractTypeErr(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expect   string
	}{
		{
			name:    "valid type error",
			message: "cannot unmarshal string into Go struct field User.age of type int",
			expect:  "field User.age should be int, not string",
		},
		{
			name:    "invalid type error format",
			message: "some other error",
			expect:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTypeErr(tt.message)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestExtractTagErr(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expect   string
	}{
		{
			name:    "max validation error",
			message: "AddError:Field validation for 'Age' failed on the 'max' tag",
			expect:  "field Age exceeds maximum value",
		},
		{
			name:    "required validation error",
			message: "AddError:Field validation for 'Name' failed on the 'required' tag",
			expect:  "field Name is required",
		},
		{
			name:    "unknown validation error",
			message: "AddError:Field validation for 'Field' failed on the 'unknown' tag",
			expect:  "field Field failed validation: unknown",
		},
		{
			name:    "invalid tag error format",
			message: "some other error",
			expect:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTagErr(tt.message)
			assert.Equal(t, tt.expect, result)
		})
	}
}