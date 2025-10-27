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
	"sync/atomic"
	"time"

	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// EngineStats holds engine load metrics
type EngineStats struct {
	Ip           string `json:"ip"`
	QueuedReqNum int32  `json:"queued_req_num"`
	PromptLength int32  `json:"prompt_length"`
	UpdatedTime  int64  `json:"updated_time"`
}

// NewEngineLoadStats creates a new EngineStats instance
func NewEngineLoadStats(ip string) *EngineStats {
	return &EngineStats{
		Ip:           ip,
		QueuedReqNum: 0,
		PromptLength: 0,
		UpdatedTime:  time.Now().UnixNano(),
	}
}

// IncrementQueuedReqNumAndPromptLength increments queue and prompt metrics
func (e *EngineStats) IncrementQueuedReqNumAndPromptLength(req *InferenceRequest, promptLength int32) {
	atomic.AddInt32(&e.QueuedReqNum, 1)
	atomic.AddInt32(&e.PromptLength, promptLength)
	e.UpdatedTime = time.Now().UnixNano()

	prom.SetLoadMetric(req.Cluster, req.Ip, e.GetQueuedReqNum(), e.GetPromptLength())
}

// DecrementQueuedReqNum decrements queue count
func (e *EngineStats) DecrementQueuedReqNum(req *InferenceRequest) {
	atomic.AddInt32(&e.QueuedReqNum, -1)
	e.UpdatedTime = time.Now().UnixNano()

	prom.SetLoadMetric(req.Cluster, req.Ip, e.GetQueuedReqNum(), e.GetPromptLength())
}

// DecrementPromptLength decrements prompt length
// Ensures single decrement by modifying req.PromptLength
func (e *EngineStats) DecrementPromptLength(req *InferenceRequest) {
	key := req.Cluster
	length := req.PromptLength
	if length <= 0 {
		logger.Debugf("DecrementPromptLength called with non-positive length: %d for request: %s", length, key)
		return
	}
	if swapped := atomic.CompareAndSwapInt32(&req.PromptLength, length, 0); !swapped {
		logger.Warnf("DecrementPromptLength failed to swap prompt length for request: %s, expected: %d, current: %d", key, length, req.PromptLength)
		return
	}
	atomic.AddInt32(&e.PromptLength, -length)
	e.UpdatedTime = time.Now().UnixNano()
	prom.SetLoadMetric(key, req.Ip, e.GetQueuedReqNum(), e.GetPromptLength())
}

// GetQueuedReqNum returns the current queued request count
func (e *EngineStats) GetQueuedReqNum() int32 {
	return atomic.LoadInt32(&e.QueuedReqNum)
}

// GetPromptLength returns the current prompt length
func (e *EngineStats) GetPromptLength() int32 {
	return atomic.LoadInt32(&e.PromptLength)
}

// MetricClean removes engine metrics for the given key
func (e *EngineStats) MetricClean(key string) {
	prom.DeleteEngineMetric(key, e.Ip)
}
