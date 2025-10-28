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
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeletePrompt_InvalidJSON(t *testing.T) {
	loadApi := LoadAPI{}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`invalid-json`)
	c.Request = httptest.NewRequest(http.MethodPost, "/delete-prompt", body)
	c.Request.Header.Set("Content-Type", "application/json")

	loadApi.DeletePrompt(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	expectedResponse := `{"status":"ERROR","error":{"code":40001400,"message":"Invalid input parameters: invalid JSON format","reason":"invalid character 'i' looking for beginning of value"},"data":null}`
	t.Logf("Response body: %s", w.Body.String())
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
