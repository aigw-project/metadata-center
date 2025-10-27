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
	"sync"
	"time"

	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// LoadStats holds all load statistics
type LoadStats struct {
	// RunningModelStats stores detailed load information for each model
	// Key: Model index using {backend_type}_{model_name} format
	// Value: ModelStats data for the corresponding model
	RunningModelStats sync.Map
	// Requests stores detailed information for each request
	// Key: RequestID
	// Value: Request details
	Requests sync.Map
}

// NewLoadStats creates a new LoadStats instance
func NewLoadStats() *LoadStats {
	return &LoadStats{}
}

// GetModelStats retrieves model statistics for the given key
func (ls *LoadStats) GetModelStats(key string) *ModelStats {
	v, ok := ls.RunningModelStats.Load(key)
	if !ok {
		return nil
	}
	return v.(*ModelStats)
}

// AddRequest adds a new inference request to load statistics
func (ls *LoadStats) AddRequest(req *InferenceRequest) {
	req.CreateTime = time.Now()
	prom.SetReplicationLatencyMillisecond(req.TimeStamp, req.RequestId)
	promptLength := req.PromptLength
	_, loaded := ls.Requests.LoadOrStore(req.RequestId, req)
	if loaded {
		logger.Infof("reqID [%s]: request ID already exists, ignoring add action", req.RequestId)
		return
	}
	key := req.Cluster
	v, loaded := ls.RunningModelStats.Load(key)
	if !loaded {
		v, loaded = ls.RunningModelStats.LoadOrStore(key, NewModelStats(key))
		if loaded {
			logger.Infof("reqID [%s]: added new model stats %s", req.RequestId, key)
		}
	}
	modelStats := v.(*ModelStats)
	engineStats := modelStats.LoadOrStore(req.Ip)
	engineStats.IncrementQueuedReqNumAndPromptLength(req, promptLength)
}

// DeleteRequest removes an inference request from load statistics
func (ls *LoadStats) DeleteRequest(req *InferenceRequest) {
	requestID := req.RequestId
	prom.SetReplicationLatencyMillisecond(req.TimeStamp, requestID)
	// Do delete first, skip delay if successful
	if ls.tryDeleteRequestStats(requestID) {
		return
	}

	// Delayed retry: handle cases where Delete occurs before Add in concurrent scenarios
	logger.Infof("reqID [%s]: request ID not found, delaying request deletion", requestID)
	time.AfterFunc(time.Second, func() {
		if !ls.tryDeleteRequestStats(requestID) {
			logger.Warnf("reqID [%s]: request ID still not found after delay, statistics may be inaccurate", requestID)
			return
		}
		logger.Infof("reqID [%s]: delayed request deletion completed", requestID)
	})
}

// tryDeleteRequestStats attempts to delete request statistics
func (ls *LoadStats) tryDeleteRequestStats(requestID string) bool {
	if v, ok := ls.Requests.LoadAndDelete(requestID); ok && v != nil {
		ls.decEngineStats(v.(*InferenceRequest))
		return true
	}

	return false
}

// decEngineStats decrements engine statistics for a request
func (ls *LoadStats) decEngineStats(req *InferenceRequest) {
	key := req.Cluster
	v, ok := ls.RunningModelStats.Load(key)
	if !ok {
		logger.Debugf("reqID [%s]: load stats cannot find model %s", req.RequestId, key)
		return
	}
	modelStats := v.(*ModelStats)
	engineStats, ok := modelStats.Load(req.Ip)
	if !ok {
		logger.Debugf("reqID [%s]: load stats cannot find engine %s on model %s", req.RequestId, req.Ip, key)
		return
	}
	engineStats.DecrementQueuedReqNum(req)
	logger.Debugf("reqID [%s]: load stats decrement queue on model %s engine %s", req.RequestId, key, req.Ip)

	// 1. Onlog phase API call: promptLength comes from cached request
	// 2. GC call: promptLength is 0 if already deleted via DELETE /api/load/prompt, otherwise original value
	engineStats.DecrementPromptLength(req)
	logger.Debugf("reqID [%s]: load stats decrement prompt length on model %s engine %s", req.RequestId, key, req.Ip)
}

// DeletePromptLength removes prompt length from statistics
func (ls *LoadStats) DeletePromptLength(req *InferenceRequest) {
	requestID := req.RequestId
	prom.SetReplicationLatencyMillisecond(req.TimeStamp, requestID)
	// Do delete first, skip delay if successful
	if ls.tryDecPromptLength(requestID) {
		return
	}

	// Delayed retry: handle cases where Delete occurs before Add in concurrent scenarios
	logger.Infof("reqID [%s]: request ID not found, delaying prompt length deletion", requestID)
	time.AfterFunc(time.Second, func() {
		if !ls.tryDecPromptLength(requestID) {
			logger.Warnf("reqID [%s]: request ID still not found after delay for prompt length deletion, may be already removed by DeleteRequest", requestID)
			return
		}
		logger.Infof("reqID [%s]: delayed prompt length deletion completed", requestID)
	})
}

// tryDecPromptLength attempts to decrement prompt length
func (ls *LoadStats) tryDecPromptLength(requestID string) bool {
	if v, ok := ls.Requests.Load(requestID); ok && v != nil {
		ls.decEnginePromptLength(v.(*InferenceRequest))
		return true
	}
	return false
}

// decEnginePromptLength decrements prompt length for engine statistics
func (ls *LoadStats) decEnginePromptLength(req *InferenceRequest) {
	key := req.Cluster
	v, ok := ls.RunningModelStats.Load(key)
	if !ok {
		logger.Debugf("reqID [%s]: load stats cannot find model %s", req.RequestId, key)
		return
	}
	modelStats := v.(*ModelStats)
	engineStats, ok := modelStats.Load(req.Ip)
	if !ok {
		logger.Debugf("reqID [%s]: load stats cannot find engine %s on model %s", req.RequestId, req.Ip, key)
		return
	}
	engineStats.DecrementPromptLength(req)
	modelStats.UpdateTime = time.Now().UnixNano()
	logger.Debugf("reqID [%s]: load stats decrement prompt length on model %s engine %s", req.RequestId, key, req.Ip)
}

// GC performs garbage collection on expired requests and statistics
func (ls *LoadStats) GC() {
	now := time.Now()
	ls.Requests.Range(func(key, value any) bool {
		req := value.(*InferenceRequest)
		if req.CreateTime.Add(requestExpireDuration).Before(now) {
			ls.Requests.Delete(key)
			ls.decEngineStats(req)
			logger.Infof("removed request from running requests, request=%v", req)
		}
		return true
	})
	nowStamps := now.UnixNano()
	expire := int64(requestExpireDuration)
	ls.RunningModelStats.Range(func(key, value any) bool {
		modelStats := value.(*ModelStats)
		if nowStamps >= modelStats.UpdateTime+expire {
			ls.RunningModelStats.Delete(key)
			modelStats.MetricClean()
			logger.Infof("removed model %s", key)
			return true
		}
		modelStats.Engines.Range(func(k, v any) bool {
			engineStats := v.(*EngineStats)
			if nowStamps >= engineStats.UpdatedTime+expire {
				// Ensure length data correctness by calling interface
				modelStats.Delete(k.(string))
				engineStats.MetricClean(key.(string))
				logger.Infof("removed engine %s on model %s", k, key)
			}
			return true
		})
		return true
	})
}
