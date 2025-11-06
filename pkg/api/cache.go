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

	"github.com/aigw-project/metadata-center/pkg/meta/cache"

	"github.com/aigw-project/metadata-center/pkg/ginx"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// CacheAPI provides cache-related HTTP endpoints
type CacheAPI struct{}

// LocationResponse represents a cache location with additional length information
type LocationResponse struct {
	IP     string `json:"ip"`     // IP address of the cache node
	Length int    `json:"length"` // Length of the cached content
}

// CacheQueryResponse contains multiple location responses for cache queries
type CacheQueryResponse struct {
	Locations []*LocationResponse `json:"locations"` // List of cache locations
}

// NewQueryResponse creates a query response from cache result mapping
func NewQueryResponse(m map[uint64]int) *CacheQueryResponse {
	r := &CacheQueryResponse{
		Locations: make([]*LocationResponse, 0, len(m)),
	}
	for key, length := range m {
		ip := cache.Decode(key)
		r.Locations = append(r.Locations, &LocationResponse{
			IP:     ip,
			Length: length * cache.DefaultChunkLen,
		})

	}
	return r
}

// Query handles cache query requests and returns matching locations
func (a *CacheAPI) Query(c *gin.Context) {
	var reqParam cache.QueryParam
	if err := ginx.ParseJSON(c, &reqParam); err != nil {
		logger.Errorf("cache api: query request error: %v", err)
		ginx.ResError(c, err)
		return
	}

	results := cache.Query(&reqParam)
	logger.Debugf("request %v got results %v", reqParam, results)

	ginx.ResSuccess(c, NewQueryResponse(results))
}

// Save handles cache save requests to store new cache entries
func (a *CacheAPI) Save(c *gin.Context) {
	var reqParam cache.SaveParam
	if err := ginx.ParseJSON(c, &reqParam); err != nil {
		logger.Errorf("cache api: save request error: %v", err)
		ginx.ResError(c, err)
		return
	}

	cache.Save(&reqParam)
	ginx.ResOK(c)
}

// CacheAdminAPI provides administrative cache management endpoints
type CacheAdminAPI struct{}

// QueryStats retrieves cache statistics and tree information
func (a *CacheAdminAPI) QueryStats(c *gin.Context) {
	var metricParam cache.ModelQueryRequest
	if err := ginx.ParseQuery(c, &metricParam); err != nil {
		logger.Errorf("cache api: query model request error: %v", err)
		ginx.ResError(c, err)
		return
	}

	ginx.ResSuccess(c, cache.ModelQuery(&metricParam))
}

// ForceGC triggers manual garbage collection on the cache
func (a *CacheAdminAPI) ForceGC(c *gin.Context) {
	cache.ForceGC()
	ginx.ResOK(c)
}
