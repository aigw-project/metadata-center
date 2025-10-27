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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/aigw-project/metadata-center/pkg/meta/load"
)

func createTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "127.0.0.1", bytes.NewBufferString(""))
	c.Request = req
	return c, w
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestAPI_InvalidParams(t *testing.T) {
	loadAPI := LoadAPI{}
	for _, f := range []func(c *gin.Context){
		loadAPI.Set,
		loadAPI.Delete,
	} {
		c, w := createTestGinContext()
		f(c)
		require.Equal(t, 400, w.Code)
	}
}

func TestLoadAPI_Params_Validate(t *testing.T) {
	loadAPI := LoadAPI{}
	t.Run("set invalid ip", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		p := load.InferenceRequest{
			Cluster:   "test",
			RequestId: "12345",
			Ip:        "invalid",
		}
		b, _ := json.Marshal(p)
		req, _ := http.NewRequest(http.MethodPost, "127.0.0.1", bytes.NewBuffer(b))
		c.Request = req
		loadAPI.Set(c)
		require.Equal(t, 400, w.Code)
	})
	t.Run("set prompt length", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		p := load.InferenceRequest{
			Cluster:      "test",
			RequestId:    "12345",
			Ip:           "1.1.1.1",
			PromptLength: -1,
		}
		b, _ := json.Marshal(p)
		req, _ := http.NewRequest(http.MethodPost, "127.0.0.1", bytes.NewBuffer(b))
		c.Request = req
		loadAPI.Set(c)
		require.Equal(t, 400, w.Code)
	})
}
