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

	"github.com/aigw-project/metadata-center/pkg/utils/errors"
	"github.com/stretchr/testify/assert"
)

func TestResponse_GetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		expect   int
	}{
		{
			name: "success status code",
			response: Response{
				Code: 200,
			},
			expect: 200,
		},
		{
			name: "error status code",
			response: Response{
				Code: 500,
			},
			expect: 500,
		},
		{
			name: "zero status code",
			response: Response{
				Code: 0,
			},
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.GetStatusCode()
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestResponseConstants(t *testing.T) {
	assert.Equal(t, "OK", OKStatus)
	assert.Equal(t, "ERROR", ErrorStatus)
}

func TestResponseStruct(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		expect   Response
	}{
		{
			name: "success response",
			response: Response{
				Code:    200,
				Status:  OKStatus,
				Data:    "test data",
				TraceID: "trace-123",
			},
			expect: Response{
				Code:    200,
				Status:  "OK",
				Data:    "test data",
				TraceID: "trace-123",
			},
		},
		{
			name: "error response",
			response: Response{
				Code:    400,
				Status:  ErrorStatus,
				Error:   errors.InvalidInput("invalid input"),
				TraceID: "trace-456",
			},
			expect: Response{
				Code:    400,
				Status:  "ERROR",
				Error:   errors.InvalidInput("invalid input"),
				TraceID: "trace-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect.Code, tt.response.Code)
			assert.Equal(t, tt.expect.Status, tt.response.Status)
			assert.Equal(t, tt.expect.Data, tt.response.Data)
			assert.Equal(t, tt.expect.TraceID, tt.response.TraceID)
			if tt.expect.Error != nil {
				assert.NotNil(t, tt.response.Error)
			}
		})
	}
}