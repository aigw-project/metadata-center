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
	"sync"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// Location represents a cache node location with IP addresses
type Location struct {
	NodeIP string `json:"node" binding:"required,ipv4"` // Node IP address
}

// Encode converts location to uint64 representation
func (l *Location) Encode() uint64 {
	return uint64(IPStr2Int(l.NodeIP)) << 32
}

// Decode converts uint64 back to location
func (l *Location) Decode(value uint64) {
	l.NodeIP = IntToIP(uint32(value >> 32))
}

// CachePool manages multiple cache instances using radix trees
type CachePool struct {
	// pool stores all cache data by domain key
	pool sync.Map
	// hash calculates prompt bytes to uint64 hashes
	hash *Hash
}

// DefaultChunkLen defines the default chunk size for hash calculation
const DefaultChunkLen = 512

// NewCachePool creates a new cache pool instance
func NewCachePool() *CachePool {
	return &CachePool{
		pool: sync.Map{},
		hash: NewHash(&HashConfig{
			ChunkLen: DefaultChunkLen,
		}),
	}
}

// QueryHash searches cache by pre-computed hash keys and returns topK matches
func (cp *CachePool) QueryHash(key string, keys []uint64, topK int) map[uint64]int {
	v, ok := cp.pool.Load(key)
	if !ok {
		return map[uint64]int{}
	}
	tree := v.(*RadixTree)
	logger.Debugf("model %s query prompt to keys: %v", key, keys)
	return tree.MatchAll(keys, topK)
}

// SaveHash stores pre-computed hash keys in cache with location information
func (cp *CachePool) SaveHash(key string, keys []uint64, local *Location) {
	v, ok := cp.pool.Load(key)
	if !ok {
		// Use LoadOrStore to avoid concurrency issues
		v, _ = cp.pool.LoadOrStore(key, NewRadixTree(key))
	}
	tree := v.(*RadixTree)
	logger.Debugf("model %s save prompt to keys: %v", key, keys)
	tree.Insert(keys, local.Encode(), nil)
}

// GC performs garbage collection on all cache trees
func (cp *CachePool) GC() {
	var wg sync.WaitGroup
	cp.pool.Range(func(key, value any) bool {
		tree := value.(*RadixTree)
		wg.Add(1)
		go func(tree *RadixTree, wg *sync.WaitGroup) {
			defer wg.Done()
			tree.CleanExpireNode()
			tree.lock.Lock()
			if len(tree.root.children) == 0 {
				cp.pool.Delete(key)
				logger.Infof("model %s cache cleaned", key)
			}
			tree.lock.Unlock()
		}(tree, &wg)
		return true
	})
	wg.Wait()
}
