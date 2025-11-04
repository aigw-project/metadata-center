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

	"github.com/stretchr/testify/assert"
)

func TestNewModelStats(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		expect    *ModelStats
	}{
		{
			name:      "create new model stats",
			modelName: "test-model",
			expect: &ModelStats{
				name: "test-model",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats(tt.modelName)
			assert.Equal(t, tt.modelName, modelStats.name)
			// Check that Engines is initialized by verifying Size() is 0
			assert.Equal(t, int32(0), modelStats.Size())
			assert.True(t, modelStats.UpdateTime > 0)
		})
	}
}

func TestModelStats_LoadOrStore(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expectIP string
	}{
		{
			name:     "load or store new engine",
			ip:       "192.168.1.1",
			expectIP: "192.168.1.1",
		},
		{
			name:     "load existing engine",
			ip:       "192.168.1.1",
			expectIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats("test-model")

			// First call should create new engine stats
			engineStats1 := modelStats.LoadOrStore(tt.ip)
			assert.Equal(t, tt.expectIP, engineStats1.Ip)
			assert.Equal(t, int32(1), modelStats.Size())

			// Second call should return same engine stats
			engineStats2 := modelStats.LoadOrStore(tt.ip)
			assert.Equal(t, engineStats1, engineStats2)
			assert.Equal(t, int32(1), modelStats.Size()) // Size should not change
		})
	}
}

func TestModelStats_Load(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		exists   bool
		expectIP string
	}{
		{
			name:     "load existing engine",
			ip:       "192.168.1.1",
			exists:   true,
			expectIP: "192.168.1.1",
		},
		{
			name:     "load non-existing engine",
			ip:       "192.168.1.2",
			exists:   false,
			expectIP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats("test-model")

			if tt.exists {
				modelStats.LoadOrStore(tt.ip)
			}

			engineStats, exists := modelStats.Load(tt.ip)
			assert.Equal(t, tt.exists, exists)
			if tt.exists {
				assert.Equal(t, tt.expectIP, engineStats.Ip)
			}
		})
	}
}

func TestModelStats_Delete(t *testing.T) {
	tests := []struct {
		name       string
		ip         string
		exists     bool
		expectSize int32
	}{
		{
			name:       "delete existing engine",
			ip:         "192.168.1.1",
			exists:     false,
			expectSize: 0,
		},
		{
			name:       "delete non-existing engine",
			ip:         "192.168.1.2",
			exists:     false,
			expectSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats("test-model")

			// Add an engine first
			modelStats.LoadOrStore("192.168.1.1")
			assert.Equal(t, int32(1), modelStats.Size())

			// Delete the specified engine
			modelStats.Delete(tt.ip)

			assert.Equal(t, tt.expectSize, modelStats.Size())

			// Verify engine is actually deleted
			_, exists := modelStats.Load(tt.ip)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

func TestModelStats_Size(t *testing.T) {
	tests := []struct {
		name    string
		engines []string
		expect  int32
	}{
		{
			name:    "empty model",
			engines: []string{},
			expect:  0,
		},
		{
			name:    "single engine",
			engines: []string{"192.168.1.1"},
			expect:  1,
		},
		{
			name:    "multiple engines",
			engines: []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
			expect:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats("test-model")

			for _, ip := range tt.engines {
				modelStats.LoadOrStore(ip)
			}

			assert.Equal(t, tt.expect, modelStats.Size())
		})
	}
}

func TestModelStats_ToEngines(t *testing.T) {
	tests := []struct {
		name    string
		engines []string
		expect  int
	}{
		{
			name:    "empty model",
			engines: []string{},
			expect:  0,
		},
		{
			name:    "single engine",
			engines: []string{"192.168.1.1"},
			expect:  1,
		},
		{
			name:    "multiple engines",
			engines: []string{"192.168.1.1", "192.168.1.2"},
			expect:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats("test-model")

			for _, ip := range tt.engines {
				modelStats.LoadOrStore(ip)
			}

			engines := modelStats.ToEngines()
			assert.Equal(t, tt.expect, len(engines))

			// Verify each engine has correct IP
			for i, engine := range engines {
				assert.Equal(t, tt.engines[i], engine.Ip)
			}
		})
	}
}

func TestModelStats_MetricClean(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "clean metrics for model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelStats := NewModelStats("test-model")

			// Should not panic
			assert.NotPanics(t, func() {
				modelStats.MetricClean()
			})
		})
	}
}
