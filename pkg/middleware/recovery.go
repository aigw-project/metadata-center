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
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/ginx"
	"github.com/aigw-project/metadata-center/pkg/utils/errors"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// Recovery creates a panic recovery middleware for Gin
// Captures and logs panics with stack traces, returns 500 error response
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		s := string(debug.Stack())
		logger.WithContext(c).Errorf("[metadata-center crash]err=%v, stack=%s\n", err, s)
		ginx.ResError(c, errors.ServerError("server error: %s", err))
	})
}
