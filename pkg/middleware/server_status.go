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
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/prom"
)

// RequestMetrics creates a middleware for collecting HTTP request metrics
// Tracks request status codes and domain-specific request counts
func RequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.Path
		if url == "/metrics" || url == "/" {
			return
		}

		status := strconv.Itoa(c.Writer.Status())
		prom.HttpRequestStatusCodeCount.WithLabelValues(url, c.Request.Method, status).Inc()
		// For GET requests, extract domain parameter and record domain instance mapping
		domain := c.Query("domain")
		if domain != "" {
			prom.DomainRequestsTotal.WithLabelValues(domain).Inc()
		}
	}
}
