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
	"sync/atomic"
	"time"

	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// ModelStats holds load information for a single model
// Each model contains multiple engines
type ModelStats struct {
	name string
	// Engines contains load information for each engine
	// Key: Engine IP
	// Value: EngineStats
	Engines sync.Map
	// Length records the current number of engines
	// Used for fast len(Engines) implementation
	Length int32
	// UpdateTime records last update time for GC
	UpdateTime int64
}

// NewModelStats creates a new ModelStats instance
func NewModelStats(name string) *ModelStats {
	return &ModelStats{
		name:       name,
		Engines:    sync.Map{},
		Length:     0,
		UpdateTime: time.Now().UnixNano(),
	}
}

// LoadOrStore loads or stores engine statistics for the given IP
func (ms *ModelStats) LoadOrStore(ip string) *EngineStats {
	v, loaded := ms.Engines.Load(ip)
	if !loaded {
		// Avoid duplicate counting, use LoadOrStore
		v, loaded = ms.Engines.LoadOrStore(ip, NewEngineLoadStats(ip))
		if !loaded {
			atomic.AddInt32(&ms.Length, 1)
			// Update metrics promptly
			prom.ModelEngineCount.WithLabelValues(ms.name).Set(float64(ms.Size()))
			logger.Infof("model %s added new engine load stats %s", ms.name, ip)
		}
	}
	ms.UpdateTime = time.Now().UnixNano()
	return v.(*EngineStats)
}

// Load retrieves engine statistics for the given IP
func (ms *ModelStats) Load(ip string) (*EngineStats, bool) {
	ms.UpdateTime = time.Now().UnixNano()
	v, ok := ms.Engines.Load(ip)
	if !ok {
		return nil, false
	}
	return v.(*EngineStats), true
}

// Delete removes engine statistics for the given IP
func (ms *ModelStats) Delete(ip string) {
	_, loaded := ms.Engines.LoadAndDelete(ip)
	// If loaded, deletion was successful
	if loaded {
		atomic.AddInt32(&ms.Length, -1)
		// Update metrics promptly
		prom.ModelEngineCount.WithLabelValues(ms.name).Set(float64(ms.Size()))
		logger.Infof("model %s deleted engine load stats %s", ms.name, ip)
	}
	ms.UpdateTime = time.Now().UnixNano()
}

// Size returns the number of engines in this model
func (ms *ModelStats) Size() int32 {
	return atomic.LoadInt32(&ms.Length)
}

// MetricClean removes metrics for this model
func (ms *ModelStats) MetricClean() {
	prom.ModelEngineCount.DeleteLabelValues(ms.name)
	prom.DeleteModelMetric(ms.name)
}

// ToEngines converts ModelStats to array format for JSON
// Cannot implement MarshalJSON: MarshalJSON passes lock by value by go vet
func (ms *ModelStats) ToEngines() []*EngineStats {
	if ms == nil {
		return nil
	}
	engines := make([]*EngineStats, 0, ms.Size())
	ms.Engines.Range(func(_, value any) bool {
		es := value.(*EngineStats)
		engines = append(engines, es)
		return true
	})
	return engines
}
