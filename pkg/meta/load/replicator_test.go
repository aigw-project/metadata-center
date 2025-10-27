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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLoadSet(t *testing.T) {
	baseReq := InferenceRequest{
		Cluster:      "test-domain",
		RequestId:    "test-request-1",
		PromptLength: 100,
		Ip:           "192.168.1.1",
	}
	basePayload, _ := json.Marshal(baseReq)

	emptyReq := InferenceRequest{}
	emptyPayload, _ := json.Marshal(emptyReq)

	tests := []struct {
		name            string
		payload         json.RawMessage
		wantErrContains string
		postCheck       func(t *testing.T)
	}{
		{
			name:    "success: add a normal inference request",
			payload: basePayload,
			postCheck: func(t *testing.T) {
				modelStats := Query(&ModelQueryRequest{Cluster: baseReq.Cluster})
				require.NotNil(t, modelStats)
				assert.Equal(t, int32(1), modelStats.Size())

				engineStats, ok := modelStats.Load(baseReq.Ip)
				require.True(t, ok)
				require.NotNil(t, engineStats)
				assert.Equal(t, int32(1), engineStats.GetQueuedReqNum())
				assert.Equal(t, int32(100), engineStats.GetPromptLength())
			},
		},
		{
			name:    "success: add an empty inference request",
			payload: emptyPayload,
			postCheck: func(t *testing.T) {
				// We just ensure it doesn't panic or error out.
				// Querying with an empty ModelInfo should show some stats.
				modelStats := Query(&ModelQueryRequest{Cluster: emptyReq.Cluster})
				require.NotNil(t, modelStats)
				assert.Equal(t, int32(1), modelStats.Size())
			},
		},
		{
			name:            "error: invalid json payload",
			payload:         json.RawMessage(`{invalid json}`),
			wantErrContains: "failed to unmarshal payload for handleLoadSet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()

			err := HandleLoadSet(tt.payload)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}

			if tt.postCheck != nil {
				tt.postCheck(t)
			}
		})
	}
}

func TestHandleLoadDelete(t *testing.T) {
	baseReq := InferenceRequest{
		Cluster:   "test-domain",
		RequestId: "test-request-2",
		Ip:        "192.168.1.2",
	}
	basePayload, _ := json.Marshal(baseReq)

	nonExistReq := InferenceRequest{
		Cluster:   "",
		RequestId: "non-exist-request",
		Ip:        "192.168.1.99",
	}
	nonExistPayload, _ := json.Marshal(nonExistReq)

	tests := []struct {
		name            string
		setup           func(t *testing.T)
		payload         json.RawMessage
		wantErrContains string
		postCheck       func(t *testing.T)
	}{
		{
			name: "success: delete an existing inference request",
			setup: func(t *testing.T) {
				require.NoError(t, HandleLoadSet(basePayload))
				modelStats := Query(&ModelQueryRequest{Cluster: baseReq.Cluster})
				engineStats, ok := modelStats.Load(baseReq.Ip)
				require.True(t, ok)
				assert.Equal(t, int32(1), engineStats.GetQueuedReqNum())
			},
			payload: basePayload,
			postCheck: func(t *testing.T) {
				modelStats := Query(&ModelQueryRequest{Cluster: baseReq.Cluster})
				engineStats, ok := modelStats.Load(baseReq.Ip)
				require.True(t, ok)
				assert.Equal(t, int32(0), engineStats.GetQueuedReqNum())
			},
		},
		{
			name:    "success: delete a non-existent request",
			payload: nonExistPayload,
		},
		{
			name:            "error: invalid json payload",
			payload:         json.RawMessage(`{invalid json}`),
			wantErrContains: "failed to unmarshal payload for handleLoadDelete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()

			if tt.setup != nil {
				tt.setup(t)
			}

			err := HandleLoadDelete(tt.payload)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}

			if tt.postCheck != nil {
				tt.postCheck(t)
			}
		})
	}
}

func TestHandleLoadPromptDelete(t *testing.T) {
	baseReq := InferenceRequest{
		Cluster:      "test-domain",
		RequestId:    "test-request-3",
		PromptLength: 300,
		Ip:           "192.168.1.3",
	}
	basePayload, _ := json.Marshal(baseReq)

	nonExistReq := InferenceRequest{
		Cluster:   "non-exist",
		RequestId: "non-exist-request",
		Ip:        "192.168.1.99",
	}
	nonExistPayload, _ := json.Marshal(nonExistReq)

	tests := []struct {
		name            string
		setup           func(t *testing.T)
		payload         json.RawMessage
		wantErrContains string
		postCheck       func(t *testing.T)
	}{
		{
			name: "success: delete prompt length from an existing request",
			setup: func(t *testing.T) {
				require.NoError(t, HandleLoadSet(basePayload))
				modelStats := Query(&ModelQueryRequest{Cluster: baseReq.Cluster})
				engineStats, ok := modelStats.Load(baseReq.Ip)
				require.True(t, ok)
				assert.Equal(t, int32(1), engineStats.GetQueuedReqNum())
				assert.Equal(t, int32(300), engineStats.GetPromptLength())
			},
			payload: basePayload,
			postCheck: func(t *testing.T) {
				modelStats := Query(&ModelQueryRequest{Cluster: baseReq.Cluster})
				engineStats, ok := modelStats.Load(baseReq.Ip)
				require.True(t, ok)
				assert.Equal(t, int32(1), engineStats.GetQueuedReqNum())
				assert.Equal(t, int32(0), engineStats.GetPromptLength())
			},
		},
		{
			name:    "success: delete prompt length from a non-existent request",
			payload: nonExistPayload,
		},
		{
			name:            "error: invalid json payload",
			payload:         json.RawMessage(`{invalid json}`),
			wantErrContains: "failed to unmarshal payload for handleLoadPromptDelete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()

			if tt.setup != nil {
				tt.setup(t)
			}

			err := HandleLoadPromptDelete(tt.payload)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}

			if tt.postCheck != nil {
				tt.postCheck(t)
			}
		})
	}
}

func TestIntegrationAllHandlers(t *testing.T) {
	Init()

	t.Run("full business flow", func(t *testing.T) {
		req := InferenceRequest{
			Cluster:      "integration-domain",
			RequestId:    "integration-request-1",
			PromptLength: 400,
			Ip:           "192.168.1.4",
		}
		payload, _ := json.Marshal(req)

		// Set request
		err := HandleLoadSet(payload)
		require.NoError(t, err)

		modelStats := Query(&ModelQueryRequest{Cluster: req.Cluster})
		engineStats, ok := modelStats.Load(req.Ip)
		require.True(t, ok)
		assert.Equal(t, int32(1), engineStats.GetQueuedReqNum())
		assert.Equal(t, int32(400), engineStats.GetPromptLength())

		// Delete prompt length
		err = HandleLoadPromptDelete(payload)
		require.NoError(t, err)

		modelStats = Query(&ModelQueryRequest{Cluster: req.Cluster})
		engineStats, ok = modelStats.Load(req.Ip)
		require.True(t, ok)
		assert.Equal(t, int32(1), engineStats.GetQueuedReqNum())
		assert.Equal(t, int32(0), engineStats.GetPromptLength())

		// Delete request
		err = HandleLoadDelete(payload)
		require.NoError(t, err)

		modelStats = Query(&ModelQueryRequest{Cluster: req.Cluster})
		engineStats, ok = modelStats.Load(req.Ip)
		require.True(t, ok)
		assert.Equal(t, int32(0), engineStats.GetQueuedReqNum())
		assert.Equal(t, int32(0), engineStats.GetPromptLength())
	})
}
