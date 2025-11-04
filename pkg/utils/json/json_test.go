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

package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalToString(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		expect string
	}{
		{
			name:   "string value",
			input:  "test string",
			expect: `"test string"`,
		},
		{
			name:   "number value",
			input:  42,
			expect: "42",
		},
		{
			name:   "boolean value",
			input:  true,
			expect: "true",
		},
		{
			name:   "struct value",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "test",
				Age:  25,
			},
			expect: `{"name":"test","age":25}`,
		},
		{
			name:   "slice value",
			input:  []string{"a", "b", "c"},
			expect: `["a","b","c"]`,
		},
		
		{
			name:   "nil value",
			input:  nil,
			expect: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarshalToString(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}