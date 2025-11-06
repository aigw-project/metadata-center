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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocation_EncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		nodeIP   string
		expected string
		hasError bool
	}{
		{
			name:     "valid IPv4 address",
			nodeIP:   "192.168.1.1",
			expected: "192.168.1.1",
			hasError: false,
		},
		{
			name:     "another valid IPv4 address",
			nodeIP:   "10.0.0.1",
			expected: "10.0.0.1",
			hasError: false,
		},
		{
			name:     "localhost IPv4",
			nodeIP:   "127.0.0.1",
			expected: "127.0.0.1",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := &Location{NodeIP: tt.nodeIP}
			encoded := loc.Encode()

			decoded := &Location{}
			decoded.Decode(encoded)

			assert.Equal(t, tt.expected, decoded.NodeIP)
		})
	}
}

func TestCachePool_QueryHash(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		keys     []uint64
		topK     int
		expected map[uint64]int
		hasError bool
	}{
		{
			name:     "non-existent key",
			key:      "non-existent",
			keys:     []uint64{123, 456},
			topK:     10,
			expected: map[uint64]int{},
			hasError: false,
		},
		{
			name:     "empty keys slice",
			key:      "test-cluster",
			keys:     []uint64{},
			topK:     5,
			expected: map[uint64]int{},
			hasError: false,
		},
		{
			name:     "zero topK value",
			key:      "test-cluster",
			keys:     []uint64{789},
			topK:     0,
			expected: map[uint64]int{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewCachePool()
			result := pool.QueryHash(tt.key, tt.keys, tt.topK)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCachePool_SaveHash(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		keys        []uint64
		location    *Location
		shouldPanic bool
	}{
		{
			name:        "save with valid location",
			key:         "test-cluster",
			keys:        []uint64{111, 222, 333},
			location:    &Location{NodeIP: "192.168.1.100"},
			shouldPanic: false,
		},
		{
			name:        "save with nil location",
			key:         "test-cluster",
			keys:        []uint64{444, 555},
			location:    nil,
			shouldPanic: true,
		},
		{
			name:        "save with empty keys",
			key:         "test-cluster",
			keys:        []uint64{},
			location:    &Location{NodeIP: "10.0.0.1"},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewCachePool()

			if tt.shouldPanic {
				// This should panic with nil location
				assert.Panics(t, func() {
					pool.SaveHash(tt.key, tt.keys, tt.location)
				})
			} else {
				// This should not panic
				pool.SaveHash(tt.key, tt.keys, tt.location)

				// Verify the data was saved by querying
				if tt.location != nil {
					result := pool.QueryHash(tt.key, tt.keys, 10)
					assert.NotNil(t, result)
				}
			}
		})
	}
}

func TestCachePool_GC(t *testing.T) {
	tests := []struct {
		name     string
		preload  map[string][]uint64 // key -> hash keys
		expected int                 // expected number of keys after GC
	}{
		{
			name:     "GC on empty pool",
			preload:  map[string][]uint64{},
			expected: 0,
		},
		{
			name: "GC on pool with data",
			preload: map[string][]uint64{
				"cluster1": {1, 2, 3},
				"cluster2": {4, 5, 6},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewCachePool()

			// Preload data
			for key, keys := range tt.preload {
				pool.SaveHash(key, keys, &Location{NodeIP: "192.168.1.1"})
			}

			// Perform GC
			pool.GC()

			// Count remaining keys
			count := 0
			pool.pool.Range(func(key, value any) bool {
				count++
				return true
			})

			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name     string
		param    *QueryParam
		expected map[uint64]int
	}{
		{
			name: "query with empty prompt hash",
			param: &QueryParam{
				Cluster: "test-cluster",
				CommonCacheParams: CommonCacheParams{
					PromptHash: []uint64{},
				},
				TopK: 10,
			},
			expected: map[uint64]int{},
		},
		{
			name: "query with nil prompt hash",
			param: &QueryParam{
				Cluster: "test-cluster",
				CommonCacheParams: CommonCacheParams{
					PromptHash: nil,
				},
				TopK: 5,
			},
			expected: map[uint64]int{},
		},
		{
			name: "query with valid prompt hash",
			param: &QueryParam{
				Cluster: "test-cluster",
				CommonCacheParams: CommonCacheParams{
					PromptHash: []uint64{123, 456},
				},
				TopK: 3,
			},
			expected: map[uint64]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize cache for testing
			Init()

			result := Query(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name     string
		param    *SaveParam
		hasError bool
	}{
		{
			name: "save with valid parameters",
			param: &SaveParam{
				Cluster: "test-cluster",
				CommonCacheParams: CommonCacheParams{
					PromptHash: []uint64{789, 101112},
				},
				Location: &Location{NodeIP: "192.168.1.200"},
			},
			hasError: false,
		},
		{
			name: "save with nil location",
			param: &SaveParam{
				Cluster: "test-cluster",
				CommonCacheParams: CommonCacheParams{
					PromptHash: []uint64{131415},
				},
				Location: nil,
			},
			hasError: true,
		},
		{
			name: "save with empty prompt hash",
			param: &SaveParam{
				Cluster: "test-cluster",
				CommonCacheParams: CommonCacheParams{
					PromptHash: []uint64{},
				},
				Location: &Location{NodeIP: "10.0.0.2"},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize cache for testing
			Init()

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					// Initialize cache for testing
					Init()

					if tt.hasError {
						// This should panic with nil location
						assert.Panics(t, func() {
							Save(tt.param)
						})
					} else {
						// This should not panic
						Save(tt.param)
					}
				})
			}
		})
	}
}

func TestModelQuery(t *testing.T) {
	tests := []struct {
		name     string
		param    *ModelQueryRequest
		expected *RadixTreeInfo
	}{
		{
			name: "query non-existent model",
			param: &ModelQueryRequest{
				Cluster: "non-existent",
			},
			expected: &RadixTreeInfo{Info: nil},
		},
		{
			name: "query with empty cluster name",
			param: &ModelQueryRequest{
				Cluster: "",
			},
			expected: &RadixTreeInfo{Info: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ModelQuery(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestForceGC(t *testing.T) {
	tests := []struct {
		name    string
		preload map[string][]uint64
	}{
		{
			name:    "force GC on empty pool",
			preload: map[string][]uint64{},
		},
		{
			name: "force GC on populated pool",
			preload: map[string][]uint64{
				"cluster1": {1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize cache for testing
			Init()

			// Preload data if needed
			for key, keys := range tt.preload {
				Save(&SaveParam{
					Cluster: key,
					CommonCacheParams: CommonCacheParams{
						PromptHash: keys,
					},
					Location: &Location{NodeIP: "192.168.1.1"},
				})
			}

			// Force GC should not panic
			ForceGC()
		})
	}
}

func TestNewCachePool(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new cache pool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewCachePool()
			assert.NotNil(t, pool)
			assert.NotNil(t, pool.hash)
		})
	}
}
