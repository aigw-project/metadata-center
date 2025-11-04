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

package ginx

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidateEitherOrTag(t *testing.T) {
	tests := []struct {
		name        string
		field1      string
		field2      string
		expectValid bool
	}{
		{
			name:        "both fields empty",
			field1:      "",
			field2:      "",
			expectValid: true,
		},
		{
			name:        "field1 set, field2 empty",
			field1:      "value1",
			field2:      "",
			expectValid: true,
		},
		{
			name:        "field1 empty, field2 set",
			field1:      "",
			field2:      "value2",
			expectValid: true,
		},
		{
			name:        "both fields set",
			field1:      "value1",
			field2:      "value2",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Field1 string `validate:"either_or=Field2"`
				Field2 string
			}

			v := validator.New()
			_ = v.RegisterValidation(eitherOrTag, validateEitherOrTag)

			testObj := TestStruct{
				Field1: tt.field1,
				Field2: tt.field2,
			}

			err := v.Struct(testObj)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateMutuallyExclusiveTag(t *testing.T) {
	tests := []struct {
		name        string
		field1      string
		field2      string
		expectValid bool
	}{
		{
			name:        "both fields empty",
			field1:      "",
			field2:      "",
			expectValid: false,
		},
		{
			name:        "field1 set, field2 empty",
			field1:      "value1",
			field2:      "",
			expectValid: true,
		},
		{
			name:        "field1 empty, field2 set",
			field1:      "",
			field2:      "value2",
			expectValid: true,
		},
		{
			name:        "both fields set",
			field1:      "value1",
			field2:      "value2",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Field1 string `validate:"mutually_exclusive=Field2"`
				Field2 string
			}

			v := validator.New()
			_ = v.RegisterValidation(mutuallyExclusiveTag, validateMutuallyExclusiveTag)

			testObj := TestStruct{
				Field1: tt.field1,
				Field2: tt.field2,
			}

			err := v.Struct(testObj)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}