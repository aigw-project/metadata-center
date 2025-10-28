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
	"regexp"
	"strings"
)

// ParseJSONFailed parses JSON binding errors and returns user-friendly error messages
// Handles syntax errors, type mismatches, and validation failures
func ParseJSONFailed(err error) error {
	reason := err.Error()

	// JSON syntax error - malformed JSON
	if strings.Contains(reason, "looking for beginning of") {
		return &ErrorInfo{
			Code:    InvalidInputCode,
			Message: fmt.Sprintf("%s: invalid JSON format", ParseJsonFieldMsg),
			Reason:  reason,
		}
	}

	// Type mismatch error - wrong data type
	errData := extractTypeErr(reason)
	if errData != "" {
		return &ErrorInfo{
			Code:    InvalidInputCode,
			Message: fmt.Sprintf("%s: %s", ParseJsonFieldMsg, errData),
			Reason:  reason,
		}
	}

	// Validation constraint violation - exceeds limits, duplicates, etc.
	errData = extractTagErr(reason)
	if errData != "" {
		return &ErrorInfo{
			Code:    InvalidInputCode,
			Message: fmt.Sprintf("%s: %s", ParseJsonFieldMsg, errData),
			Reason:  reason,
		}
	}

	return &ErrorInfo{
		Code:    InvalidInputCode,
		Message: ParseJsonFieldMsg,
		Reason:  reason,
	}
}

// typeRegexp matches type mismatch errors from JSON unmarshaling
var typeRegexp = regexp.MustCompile(`cannot unmarshal (.+?) into Go struct field (.+?) of type (.+?)$`)

// extractTypeErr extracts type mismatch information from error message
// Returns formatted error string for type validation failures
func extractTypeErr(message string) string {
	match := typeRegexp.FindStringSubmatch(message)
	if len(match) <= 3 {
		return ""
	}
	return fmt.Sprintf("field %s should be %s, not %s", match[2], match[3], match[1])
}

// tagErrMap maps validation tag names to user-friendly error messages
var tagErrMap = map[string]string{
	"max":         "exceeds maximum value",
	"min":         "below minimum value",
	"lte":         "exceeds maximum value",
	"gte":         "below minimum value",
	"required":    "is required",
	"required_if": "is required",
	"lowercase":   "only lowercase characters allowed",
	"alpha":       "invalid input",                // Only ASCII letters allowed (no numbers, punctuation, control chars)
	"oneof":       "invalid input",                // Enum value validation
	"unique":      "duplicate values not allowed", // Duplicate values not allowed
	"fqdn":        "invalid domain",
	"ip":          "invalid IP address",
}

// tagRegexp matches validation tag errors from field validation
var tagRegexp = regexp.MustCompile(`AddError:Field validation for '(.+?)' failed on the '(.+?)' tag$`)

// extractTagErr extracts validation tag information from error message
// Returns formatted error string for validation constraint failures
func extractTagErr(message string) string {
	match := tagRegexp.FindStringSubmatch(message)
	if len(match) <= 2 {
		return ""
	}

	tagErr, exist := tagErrMap[match[2]]
	if !exist { // Unmapped tag
		return fmt.Sprintf("field %s failed validation: %s", match[1], match[2])
	}

	return fmt.Sprintf("field %s %s", match[1], tagErr)
}
