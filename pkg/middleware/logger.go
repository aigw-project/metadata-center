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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/aigw-project/metadata-center/pkg/config"
	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

// Logger creates a Gin middleware for HTTP request logging
func Logger() gin.HandlerFunc {
	c := config.C.Log

	var output io.Writer
	switch c.GinOutput {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		if name := c.GinOutputFile; name != "" {
			_ = os.MkdirAll(filepath.Dir(name), 0777)

			f, err := rotatelogs.New(name+".%Y-%m-%d-%H",
				rotatelogs.WithLinkName(name),
				rotatelogs.WithRotationTime(time.Duration(c.RotationTime)*time.Hour),
				rotatelogs.WithRotationCount(uint(c.RotationCount)))
			if err != nil {
				return nil
			}
			output = f
		}
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    output,
		Formatter: customLogFormatter,
	})
}

// customLogFormatter formats HTTP request logs with extended fields
func customLogFormatter(param gin.LogFormatterParams) string {
	traceID, exists := param.Keys[string(trace.TraceKey)]
	if !exists {
		traceID = ""
	}
	userId, exists := param.Keys["userId"]
	if !exists {
		userId = ""
	}
	reqID, exists := param.Keys["requestId"]
	if !exists {
		reqID = ""
	}
	eventType, exists := param.Keys["eventType"]
	if !exists {
		eventType = ""
	}

	path := param.Request.URL.Path
	if path != "/metrics" && path != "/" {
		prom.HttpRequestDuration.WithLabelValues(param.Method, param.Request.URL.Path).Observe(float64(param.Latency.Microseconds()))
	}

	return fmt.Sprintf("reqTime=%s, traceID=[%s], requestID=[%s], userID=%s, clientIP=%s, method=%s, path=%s, url=%s, eventType=%s, proto=%s, respCode=%d, latency=%dus, UA=%s, respBodySize=%d，err=%s\n",
		param.TimeStamp.Format(time.RFC3339Nano),
		traceID,
		reqID,
		userId,
		param.ClientIP,
		param.Method,
		param.Request.URL.Path,
		param.Path,
		eventType,
		param.Request.Proto,
		param.StatusCode,
		param.Latency.Microseconds(),
		param.Request.UserAgent(),
		param.BodySize,
		param.ErrorMessage,
	)
}
