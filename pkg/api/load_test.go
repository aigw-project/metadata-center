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

package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/aigw-project/metadata-center/pkg/meta/load"
	"github.com/aigw-project/metadata-center/pkg/replicator"
)

func init() {
	os.Setenv("POD_IP", "127.0.0.1")
	load.Init()
	replicator.Init()
}

func TestLoadAPI(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		expectedStatus int
		expectedBody   string
	}{
		// Query tests
		{
			name:           "Query_Success",
			method:         http.MethodGet,
			path:           "/query?cluster=test-cluster",
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Query_InvalidQueryParams",
			method:         http.MethodGet,
			path:           "/query?invalid=param",
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
		// Set tests
		{
			name:           "Set_Success",
			method:         http.MethodPost,
			path:           "/set",
			body:           `{"cluster":"test-cluster","request_id":"test-request","ip":"192.168.1.1"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Set_InvalidJSON",
			method:         http.MethodPost,
			path:           "/set",
			body:           "invalid-json",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		// Delete tests
		{
			name:           "Delete_Success",
			method:         http.MethodDelete,
			path:           "/delete",
			body:           `{"request_id":"test-request"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete_InvalidJSON",
			method:         http.MethodDelete,
			path:           "/delete",
			body:           "invalid-json",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		// DeletePrompt tests
		{
			name:           "DeletePrompt_Success",
			method:         http.MethodDelete,
			path:           "/delete-prompt",
			body:           `{"request_id":"test-request"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DeletePrompt_InvalidJSON",
			method:         http.MethodPost,
			path:           "/delete-prompt",
			body:           "invalid-json",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":{"code":40001400,"message":"Invalid input parameters: invalid JSON format","reason":"invalid character 'i' looking for beginning of value"},"data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadApi := LoadAPI{}
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body *bytes.Buffer
			if tt.body != "" {
				body = bytes.NewBufferString(tt.body)
			} else {
				body = bytes.NewBufferString("")
			}

			c.Request = httptest.NewRequest(tt.method, tt.path, body)
			if tt.contentType != "" {
				c.Request.Header.Set("Content-Type", tt.contentType)
			}

			switch tt.method {
			case http.MethodGet:
				loadApi.Query(c)
			case http.MethodPost:
				if tt.path == "/set" {
					loadApi.Set(c)
				} else {
					loadApi.DeletePrompt(c)
				}
			case http.MethodDelete:
				if tt.path == "/delete" {
					loadApi.Delete(c)
				} else {
					loadApi.DeletePrompt(c)
				}
			}

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				t.Logf("Response body: %s", w.Body.String())
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}