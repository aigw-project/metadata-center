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
	"encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/utils/errors"
	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

// ParseJSON Parse body json data to struct
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.ParseJSONFailed(err)
	}
	return nil
}

// ParseQuery Parse query parameter to struct
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.InvalidInput("Parse request query failed: %s", err.Error())
	}
	return nil
}

// ResOK Response success with prom ok
func ResOK(c *gin.Context) {
	ResJSON(c, Response{
		Code:    200,
		Status:  OKStatus,
		TraceID: GetTraceID(c),
	})
}

// ResSuccess Response data object
func ResSuccess(c *gin.Context, v interface{}) {
	ResJSON(c, Response{
		Code:    200,
		Status:  OKStatus,
		Data:    v,
		TraceID: GetTraceID(c),
	})
}

// ResJSON Response json data with prom code
func ResJSON(c *gin.Context, response Response) {
	buf, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	// c.Set(ResBodyKey, buf)
	c.Data(response.GetStatusCode(), "application/json; charset=utf-8", buf)
	c.Abort()
}

// ResError Response error object and parse error prom code
func ResError(c *gin.Context, err error) {
	var resErr *errors.ErrorInfo
	if err != nil {
		if e, ok := err.(*errors.ErrorInfo); ok {
			resErr = e
		} else {
			resErr = errors.ServerError("server error: %s", err.Error())
		}
	} else {
		resErr = errors.ServerError("")
	}

	_ = c.Error(resErr)
	ResJSON(c, Response{
		Code:    resErr.GetStatusCode(),
		Status:  ErrorStatus,
		Error:   resErr,
		TraceID: GetTraceID(c),
	})
}

// GetTraceID retrieves trace ID from gin context
func GetTraceID(c *gin.Context) string {
	traceID, _ := c.Get(string(trace.TraceKey))
	if ret, ok := traceID.(string); ok {
		return ret
	}
	return ""
}
