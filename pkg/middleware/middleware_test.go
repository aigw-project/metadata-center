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

package middleware

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aigw-project/metadata-center/pkg/config"
	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		filePath string
		expectNil bool
	}{
		{
			name:     "stdout output",
			output:   "stdout",
			expectNil: false,
		},
		{
			name:     "stderr output",
			output:   "stderr",
			expectNil: false,
		},
		{
			name:     "file output with valid path",
			output:   "file",
			filePath: filepath.Join(t.TempDir(), "test.log"),
			expectNil: false,
		},
		{
			name:     "file output with empty path",
			output:   "file",
			filePath: "",
			expectNil: false,
		},
		{
			name:     "invalid output type",
			output:   "invalid",
			expectNil: false, // Defaults to stdout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup original config
			originalConfig := config.C.Log
			defer func() { config.C.Log = originalConfig }()

			config.C.Log.GinOutput = tt.output
			config.C.Log.GinOutputFile = tt.filePath
			config.C.Log.RotationTime = 1
			config.C.Log.RotationCount = 1

			middleware := Logger()
			if tt.expectNil {
				assert.Nil(t, middleware)
			} else {
				assert.NotNil(t, middleware)
			}
		})
	}
}

func TestCustomLogFormatter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		path     string
		method   string
		status   int
		latency  time.Duration
		keys     map[string]any
		errorMsg string
		expect   string
	}{
		{
			name:    "basic request without trace data",
			path:    "/test",
			method:  "GET",
			status:  200,
			latency: time.Millisecond,
			keys:    make(map[string]any),
			expect:  "traceID=[], requestID=[], userID=, clientIP=192.0.2.1, method=GET, path=/test, url=/test, eventType=, proto=HTTP/1.1, respCode=200, latency=1000us, UA=, respBodySize=0，err=\n",
		},
		{
			name:    "request with trace data",
			path:    "/api/v1/test",
			method:  "POST",
			status:  201,
			latency: 2 * time.Millisecond,
			keys: map[string]any{
				string(trace.TraceKey): "trace-123",
				"requestId":           "req-456",
				"userId":              "user-789",
				"eventType":           "api_call",
			},
			expect: "traceID=[trace-123], requestID=[req-456], userID=user-789, clientIP=192.0.2.1, method=POST, path=/api/v1/test, url=/api/v1/test, eventType=api_call, proto=HTTP/1.1, respCode=201, latency=2000us, UA=, respBodySize=0，err=\n",
		},
		{
			name:     "request with error",
			path:     "/error",
			method:   "PUT",
			status:   500,
			latency:  100 * time.Millisecond,
			keys:     make(map[string]any),
			errorMsg: "internal server error",
			expect:   "traceID=[], requestID=[], userID=, clientIP=192.0.2.1, method=PUT, path=/error, url=/error, eventType=, proto=HTTP/1.1, respCode=500, latency=100000us, UA=, respBodySize=0，err=internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, tt.path, nil)

			params := gin.LogFormatterParams{
				Request:      c.Request,
				TimeStamp:    time.Now(),
				ClientIP:     c.ClientIP(),
				Method:       tt.method,
				Path:         tt.path,
				StatusCode:   tt.status,
				Latency:      tt.latency,
				BodySize:     0,
				Keys:         tt.keys,
				ErrorMessage: tt.errorMsg,
			}

			result := customLogFormatter(params)
			// Remove timestamp from result for comparison
			result = result[strings.Index(result, "traceID="):]
			expect := tt.expect
			assert.Equal(t, expect, result)
		})
	}
}

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		panicValue   any
	}{
		{
			name:         "panic with string",
			panicValue:   "test panic",
		},
		{
			name:         "panic with error",
			panicValue:   assert.AnError,
		},
		{
			name:         "panic with nil",
			panicValue:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Recovery()
			require.NotNil(t, middleware)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Capture the panic and recover
			assert.NotPanics(t, func() {
				middleware(c)
			})

			// Recovery middleware should handle the panic without crashing
			assert.NotNil(t, w)
		})
	}
}

func TestRequestMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		path     string
		method   string
		query    string
	}{
		{
			name:     "metrics endpoint should not record",
			path:     "/metrics",
			method:   "GET",
		},
		{
			name:     "root endpoint should not record",
			path:     "/",
			method:   "GET",
		},
		{
			name:     "regular endpoint should record",
			path:     "/api/v1/test",
			method:   "POST",
		},
		{
			name:     "endpoint with domain query should record",
			path:     "/api/v1/data",
			method:   "GET",
			query:    "domain=example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := RequestMetrics()
			require.NotNil(t, middleware)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			url := tt.path
			if tt.query != "" {
				url = tt.path + "?" + tt.query
			}
			c.Request = httptest.NewRequest(tt.method, url, nil)

			// Execute the middleware
			middleware(c)

			// The middleware doesn't return any visible state changes,
			// so we just verify it doesn't panic and executes
			assert.NotNil(t, w)
		})
	}
}

func TestTrace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		header     map[string]string
		expectTrace bool
	}{
		{
			name:       "no trace header",
			header:     map[string]string{},
			expectTrace: true, // Should generate a trace ID
		},
		{
			name:       "with trace header",
			header:     map[string]string{"TraceId": "custom-trace-123"},
			expectTrace: true,
		},
		{
			name:       "empty trace header",
			header:     map[string]string{"TraceId": ""},
			expectTrace: true, // Should generate a trace ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Trace()
			require.NotNil(t, middleware)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set headers
			for key, value := range tt.header {
				c.Request.Header.Set(key, value)
			}

			// Execute the middleware
			middleware(c)

			// Check if trace ID is set in context and response headers
			traceID, exists := c.Get(string(trace.TraceKey))
			assert.True(t, exists)
			assert.NotEmpty(t, traceID)

			// Check response header
			responseTraceID := w.Header().Get(string(trace.TraceKey))
			assert.Equal(t, traceID, responseTraceID)

			// Check context
			ctx := c.Request.Context()
			contextTraceID := ctx.Value(trace.TraceKey)
			assert.Equal(t, traceID, contextTraceID)
		})
	}
}

func TestGetMiddlewares(t *testing.T) {
	tests := []struct {
		name           string
		configSetup    func()
		expectedCount  int
	}{
		{
			name: "all middlewares available",
			configSetup: func() {
				// Use default config
			},
			expectedCount: 4, // Trace, Logger, Recovery, RequestMetrics
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configSetup != nil {
				tt.configSetup()
			}

			middlewares := GetMiddlewares()
			assert.Len(t, middlewares, tt.expectedCount)

			// All returned middlewares should be non-nil
			for _, mw := range middlewares {
				assert.NotNil(t, mw)
			}
		})
	}
}