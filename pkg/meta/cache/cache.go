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

package cache

import (
	"time"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// cachePool is the global cache pool instance
var cachePool *CachePool

// cronClean performs periodic garbage collection on the cache pool
func cronClean(pool *CachePool, ticker *time.Ticker) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("cache init goroutine panic: %v", r)
		}
		ticker.Stop()
		logger.Errorf("cache init goroutine exited")
	}()
	for range ticker.C {
		start := time.Now()
		pool.GC()
		logger.Infof("completed full pool GC, duration: %d", time.Since(start))
	}
}

// Init initializes the cache system with background garbage collection
func Init() {
	cachePool = NewCachePool()
	go func() {
		ticker := time.NewTicker(gcInterval)
		cronClean(cachePool, ticker)
	}()
	logger.Infof("cache initialization successful")
}

// CommonCacheParams contains common parameters for cache operations
type CommonCacheParams struct {
	PromptHash []uint64 `json:"prompt_hash" binding:"required"`
}

// QueryParam defines parameters for cache query operations
type QueryParam struct {
	Cluster string `json:"cluster" form:"cluster" binding:"required"`
	CommonCacheParams
	// TopK represents the maximum number of results to return, 0 means use default configuration value
	TopK int `json:"top_k"`
}

// Query searches the cache for matching entries based on prompt or prompt hash
func Query(p *QueryParam) map[uint64]int {
	if len(p.PromptHash) != 0 {
		return cachePool.QueryHash(p.Cluster, p.PromptHash, p.TopK)
	}
	return map[uint64]int{}
}

// SaveParam defines parameters for cache save operations
type SaveParam struct {
	Cluster string `json:"cluster" form:"cluster" binding:"required"`
	CommonCacheParams
	Location *Location `json:"location" binding:"required"`
}

// Save stores a new cache entry with the specified parameters
func Save(p *SaveParam) {
	if len(p.PromptHash) != 0 {
		cachePool.SaveHash(p.Cluster, p.PromptHash, p.Location)
		return
	}
}

// ModelQueryRequest defines parameters for model tree information queries
type ModelQueryRequest struct {
	Cluster string `json:"cluster" form:"cluster" binding:"required"`
}

// ModelQuery retrieves radix tree information for a specific cluster
func ModelQuery(p *ModelQueryRequest) *RadixTreeInfo {
	// Check if cachePool is initialized
	if cachePool == nil {
		return &RadixTreeInfo{}
	}

	v, ok := cachePool.pool.Load(p.Cluster)
	if !ok {
		return &RadixTreeInfo{}
	}
	tree := v.(*RadixTree)
	return tree.GetTreeInfo()
}

// ForceGC triggers immediate garbage collection on the cache pool
func ForceGC() {
	logger.Infof("cache trigger force gc")
	cachePool.GC()
}
