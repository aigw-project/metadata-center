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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEngineLoadStats(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expectIP string
	}{
		{
			name:     "valid IP",
			ip:       "192.168.1.1",
			expectIP: "192.168.1.1",
		},
		{
			name:     "empty IP",
			ip:       "",
			expectIP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := NewEngineLoadStats(tt.ip)
			assert.Equal(t, tt.expectIP, engineStats.Ip)
			assert.Equal(t, int32(0), engineStats.QueuedReqNum)
			assert.Equal(t, int32(0), engineStats.PromptLength)
			assert.True(t, engineStats.UpdatedTime > 0)
		})
	}
}

func TestEngineStats_IncrementQueuedReqNumAndPromptLength(t *testing.T) {
	tests := []struct {
		name         string
		initialReq   int32
		initialPrompt int32
		promptLength int32
		expectReq    int32
		expectPrompt int32
	}{
		{
			name:         "increment from zero",
			initialReq:   0,
			initialPrompt: 0,
			promptLength: 100,
			expectReq:    1,
			expectPrompt: 100,
		},
		{
			name:         "increment from existing values",
			initialReq:   5,
			initialPrompt: 500,
			promptLength: 200,
			expectReq:    6,
			expectPrompt: 700,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := &EngineStats{
				Ip:           "192.168.1.1",
				QueuedReqNum: tt.initialReq,
				PromptLength: tt.initialPrompt,
				UpdatedTime:  time.Now().UnixNano(),
			}

			req := &InferenceRequest{
				Cluster: "test-cluster",
				Ip:      "192.168.1.1",
			}

			engineStats.IncrementQueuedReqNumAndPromptLength(req, tt.promptLength)

			assert.Equal(t, tt.expectReq, engineStats.GetQueuedReqNum())
			assert.Equal(t, tt.expectPrompt, engineStats.GetPromptLength())
			assert.True(t, engineStats.UpdatedTime > 0)
		})
	}
}

func TestEngineStats_DecrementQueuedReqNum(t *testing.T) {
	tests := []struct {
		name         string
		initialReq   int32
		initialPrompt int32
		expectReq    int32
	}{
		{
			name:         "decrement from one",
			initialReq:   1,
			initialPrompt: 100,
			expectReq:    0,
		},
		{
			name:         "decrement from multiple",
			initialReq:   5,
			initialPrompt: 500,
			expectReq:    4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := &EngineStats{
				Ip:           "192.168.1.1",
				QueuedReqNum: tt.initialReq,
				PromptLength: tt.initialPrompt,
				UpdatedTime:  time.Now().UnixNano(),
			}

			req := &InferenceRequest{
				Cluster: "test-cluster",
				Ip:      "192.168.1.1",
			}

			engineStats.DecrementQueuedReqNum(req)

			assert.Equal(t, tt.expectReq, engineStats.GetQueuedReqNum())
			assert.Equal(t, tt.initialPrompt, engineStats.GetPromptLength())
			assert.True(t, engineStats.UpdatedTime > 0)
		})
	}
}

func TestEngineStats_DecrementPromptLength(t *testing.T) {
	tests := []struct {
		name           string
		initialPrompt  int32
		promptLength   int32
		expectPrompt   int32
		expectSwapped  bool
	}{
		{
			name:          "decrement positive length",
			initialPrompt: 100,
			promptLength:  50,
			expectPrompt:  50,
			expectSwapped: true,
		},
		{
			name:          "decrement zero length",
			initialPrompt: 100,
			promptLength:  0,
			expectPrompt:  100,
			expectSwapped: false,
		},
		{
			name:          "decrement negative length",
			initialPrompt: 100,
			promptLength:  -10,
			expectPrompt:  100,
			expectSwapped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := &EngineStats{
				Ip:           "192.168.1.1",
				QueuedReqNum: 1,
				PromptLength: tt.initialPrompt,
				UpdatedTime:  time.Now().UnixNano(),
			}

			req := &InferenceRequest{
				Cluster:      "test-cluster",
				Ip:           "192.168.1.1",
				PromptLength: tt.promptLength,
			}

			engineStats.DecrementPromptLength(req)

			assert.Equal(t, tt.expectPrompt, engineStats.GetPromptLength())
			assert.True(t, engineStats.UpdatedTime > 0)
			
			if tt.expectSwapped {
				assert.Equal(t, int32(0), req.PromptLength)
			} else {
				assert.Equal(t, tt.promptLength, req.PromptLength)
			}
		})
	}
}

func TestEngineStats_GetQueuedReqNum(t *testing.T) {
	tests := []struct {
		name       string
		queuedReqNum int32
		expect     int32
	}{
		{
			name:       "zero requests",
			queuedReqNum: 0,
			expect:     0,
		},
		{
			name:       "multiple requests",
			queuedReqNum: 5,
			expect:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := &EngineStats{
				QueuedReqNum: tt.queuedReqNum,
			}

			result := engineStats.GetQueuedReqNum()
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestEngineStats_GetPromptLength(t *testing.T) {
	tests := []struct {
		name         string
		promptLength int32
		expect       int32
	}{
		{
			name:         "zero length",
			promptLength: 0,
			expect:       0,
		},
		{
			name:         "positive length",
			promptLength: 1000,
			expect:       1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := &EngineStats{
				PromptLength: tt.promptLength,
			}

			result := engineStats.GetPromptLength()
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestEngineStats_MetricClean(t *testing.T) {
	tests := []struct {
		name  string
		ip    string
		key   string
	}{
		{
			name: "clean metrics for engine",
			ip:   "192.168.1.1",
			key:  "test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engineStats := &EngineStats{
				Ip: tt.ip,
			}

			// Should not panic
			assert.NotPanics(t, func() {
				engineStats.MetricClean(tt.key)
			})
		})
	}
}