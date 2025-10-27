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

package load

import (
	"time"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// ModelQueryRequest represents a query request for model statistics
type ModelQueryRequest struct {
	Cluster string `json:"cluster" binding:"required" form:"cluster"`
}

// InferenceRequest represents an inference request with load metrics
type InferenceRequest struct {
	Cluster      string    `json:"cluster" binding:"required" form:"cluster"`
	RequestId    string    `json:"request_id" binding:"required" form:"request_id"`
	PromptLength int32     `json:"prompt_length,omitempty" binding:"gte=0"`
	Ip           string    `json:"ip" binding:"required,ipv4" form:"ip"`
	TimeStamp    int64     `json:"timestamp,omitempty" form:"timestamp"`
	CreateTime   time.Time `json:"-"`
}

// DeletionInferenceRequest represents an inference request for deletion
type DeletionInferenceRequest struct {
	RequestId string `json:"request_id" binding:"required" form:"request_id"`
	TimeStamp int64  `json:"timestamp,omitempty" form:"timestamp"`
}

var loadStats *LoadStats

// Init initializes the load statistics system
func Init() {
	loadStats = NewLoadStats()
	go func() {
		ticker := time.NewTicker(gcInterval)
		cronClean(ticker, loadStats)
	}()
	logger.Infof("initializing metadata: load process")
}

// cronClean runs periodic garbage collection for load statistics
func cronClean(ticker *time.Ticker, stats *LoadStats) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("load initialization goroutine panicked: %v", r)
		}
		ticker.Stop()
		logger.Errorf("load initialization goroutine exited")
	}()

	for range ticker.C {
		start := time.Now()
		stats.GC()
		logger.Infof("completed full load stats garbage collection, duration: %d", time.Since(start))
	}
}

// Query retrieves model statistics for the given cluster
func Query(req *ModelQueryRequest) *ModelStats {
	return loadStats.GetModelStats(req.Cluster)
}

// Set adds a new inference request to load statistics
func Set(req *InferenceRequest) {
	loadStats.AddRequest(req)
}

// Delete removes an inference request from load statistics
func Delete(req *DeletionInferenceRequest) {
	loadStats.DeleteRequest(req)
}

// PromptDelete removes prompt length from statistics
func PromptDelete(req *DeletionInferenceRequest) {
	loadStats.DeletePromptLength(req)
}
