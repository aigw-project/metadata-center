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
	"context"

	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

// Trace creates a middleware for handling request tracing
// Extracts or generates trace ID from headers and sets it in context and response headers
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("TraceId")
		if traceID == "" {
			traceID = trace.TraceID()
		}
		c.Set(string(trace.TraceKey), traceID)

		ctx := context.WithValue(c.Request.Context(), trace.TraceKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set(string(trace.TraceKey), traceID)
	}
}
