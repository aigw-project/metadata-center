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
	"container/heap"
	"encoding/binary"
	"net"
)

// longestCommonPrefix returns the longest common prefix of two keys
func longestCommonPrefix(a, b []uint64) int {
	l := min(len(a), len(b))
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return l
}

// sliceEqual compares if two prefix slices are equal
func sliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// IPStr2Int converts IPv4 format string to big-endian uint32
func IPStr2Int(s string) uint32 {
	ip := net.ParseIP(s)
	if len(ip) == 0 {
		return 0
	}
	b := []byte(net.ParseIP(s).To4())
	return binary.BigEndian.Uint32(b)
}

// IntToIP converts a big-endian uint32 to IPv4 format string
func IntToIP(v uint32) string {
	if v == 0 {
		return ""
	}
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return net.IP(b).String()
}

// MinHeapEntry represents an entry in the min-heap with key-value pair and index
type MinHeapEntry struct {
	index int    // Heap index position
	key   uint64 // Key identifier
	value int    // Value for comparison
}

// MinHeap implements heap.Interface for maintaining top-K elements
type MinHeap []*MinHeapEntry

// Len returns the number of elements in the heap
func (h MinHeap) Len() int { return len(h) }

// Less compares two elements for heap ordering (min-heap)
func (h MinHeap) Less(i, j int) bool { return h[i].value < h[j].value }

// Swap exchanges elements and updates their indices
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

// Push adds an element to the heap
func (h *MinHeap) Push(x any) {
	n := len(*h)
	item := x.(*MinHeapEntry)
	item.index = n
	*h = append(*h, item)
}

// Pop removes and returns the smallest element from the heap
func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1 // Mark as removed
	*h = old[0 : n-1]
	return item
}

// TopKMap maintains top-K largest values using a min-heap
// TODO: Optimization potential - may not need min-heap directly due to RadixTree patterns
type TopKMap struct {
	k       int                      // Maximum number of values to store
	m       map[uint64]*MinHeapEntry // Map of key to heap entries
	minHeap *MinHeap                 // Min-heap for maintaining top-K
}

// NewTopKMap creates a new TopKMap with specified capacity
func NewTopKMap(k int) *TopKMap {
	if k <= 0 {
		k = defaultQueryTopK
	}
	return &TopKMap{
		k:       k,
		m:       map[uint64]*MinHeapEntry{},
		minHeap: &MinHeap{},
	}
}

// Add inserts or updates a key-value pair, maintaining top-K order
// No locking - each request constructs TopK independently, no concurrent updates
func (m *TopKMap) Add(key uint64, value int) {
	if entry, ok := m.m[key]; ok {
		if value > entry.value {
			entry.value = value
			heap.Fix(m.minHeap, entry.index)
		}
		return
	}
	if m.minHeap.Len() < m.k {
		entry := &MinHeapEntry{
			key:   key,
			value: value,
		}
		heap.Push(m.minHeap, entry)
		m.m[key] = entry
		return
	}
	// Replace smallest element if new value is larger
	topEntry := (*m.minHeap)[0]
	if value > topEntry.value {
		delete(m.m, topEntry.key)
		topEntry.key = key
		topEntry.value = value
		m.m[key] = topEntry
		heap.Fix(m.minHeap, 0)
	}
}

// GetMap returns the current top-K entries as a map
func (m *TopKMap) GetMap() map[uint64]int {
	ret := make(map[uint64]int, len(m.m))
	for key, entry := range m.m {
		ret[key] = entry.value
	}
	return ret
}

// Encode converts location to uint64 representation
func Encode(ip string) uint64 {
	return uint64(IPStr2Int(ip)) << 32
}

// Decode converts uint64 back to location
func Decode(value uint64) string {
	return IntToIP(uint32(value >> 32))
}
