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

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aigw-project/metadata-center/pkg/api"
	"github.com/aigw-project/metadata-center/pkg/log"
	"github.com/aigw-project/metadata-center/pkg/replicator"
)

// RegisterLoadAPI registers load-related API endpoints
// Includes stats and prompt management endpoints
func RegisterLoadAPI(g *gin.RouterGroup) {
	loadAPI := api.LoadAPI{}
	gGroup := g.Group("/v1/load")
	stats := gGroup.Group("stats")
	{
		stats.GET("", loadAPI.Query)
		stats.POST("", loadAPI.Set)
		stats.DELETE("", loadAPI.Delete)
	}
	prompt := gGroup.Group("prompt")
	{
		prompt.DELETE("", loadAPI.DeletePrompt)
	}
}

// RegisterStatusAPI registers metrics endpoint for Prometheus
func RegisterStatusAPI(g *gin.RouterGroup) {
	gGroup := g.Group("/metrics")
	{
		gGroup.GET("", gin.WrapH(promhttp.Handler()))
	}
}

// RegisterLogAPI registers log management endpoints
func RegisterLogAPI(g *gin.RouterGroup) {
	gGroup := g.Group("/log")
	{
		gGroup.POST("level", log.UpdateLogLevel)
	}
}

// RegisterReplicateAPI registers replication event endpoints
func RegisterReplicateAPI(g *gin.RouterGroup) {
	gGroup := g.Group("/v1/replica/event")
	{
		gGroup.POST("", replicator.HandleReplicateEvent)
	}
}
