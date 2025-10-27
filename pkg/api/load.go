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
	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/ginx"
	"github.com/aigw-project/metadata-center/pkg/meta/load"
	"github.com/aigw-project/metadata-center/pkg/replicator"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// RequestIdCtxKey is the context key for storing request ID
const RequestIdCtxKey = "requestId"

// LoadAPI handles load statistics related HTTP endpoints
type LoadAPI struct {
}

// Query handles GET requests for querying model load statistics
func (a *LoadAPI) Query(c *gin.Context) {
	var metricParam load.ModelQueryRequest
	if err := ginx.ParseQuery(c, &metricParam); err != nil {
		logger.Errorf("load api: query model request error: %v", err)
		ginx.ResError(c, err)
		return
	}

	stat := load.Query(&metricParam)
	ginx.ResSuccess(c, stat.ToEngines())
}

// Set handles POST requests for setting load statistics
func (a *LoadAPI) Set(c *gin.Context) {
	var reqParam load.InferenceRequest
	if err := ginx.ParseJSON(c, &reqParam); err != nil {
		logger.Errorf("load api: set request error: %v", err)
		ginx.ResError(c, err)
		return
	}

	c.Set(RequestIdCtxKey, reqParam.RequestId)
	load.Set(&reqParam)
	replicator.Replicate(c, load.LoadStatsSet, reqParam) // Replicate to other instances

	ginx.ResOK(c)
}

// Delete handles DELETE requests for removing load statistics
func (a *LoadAPI) Delete(c *gin.Context) {
	var reqParam load.DeletionInferenceRequest
	if err := ginx.ParseJSON(c, &reqParam); err != nil {
		logger.Errorf("load api: delete request error: %v", err)
		ginx.ResError(c, err)
		return
	}

	c.Set(RequestIdCtxKey, reqParam.RequestId)
	load.Delete(&reqParam)
	replicator.Replicate(c, load.LoadStatsDelete, reqParam) // Replicate to other instances

	ginx.ResOK(c)
}

// DeletePrompt handles DELETE requests for removing prompt length statistics
func (a *LoadAPI) DeletePrompt(c *gin.Context) {
	var reqParam load.DeletionInferenceRequest
	if err := ginx.ParseJSON(c, &reqParam); err != nil {
		logger.Errorf("load api: delete request prompt length error: %v", err)
		ginx.ResError(c, err)
		return
	}

	c.Set(RequestIdCtxKey, reqParam.RequestId)
	load.PromptDelete(&reqParam)
	replicator.Replicate(c, load.LoadPromptDelete, reqParam) // Replicate to other instances

	ginx.ResOK(c) // Return success response
}
