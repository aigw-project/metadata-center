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
	"math/rand"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

func TestMain(m *testing.M) {
	logger.SetLevel(logger.DebugLevel)
	os.Exit(m.Run())
}

// TestRadixTree tests basic radix tree functionality including insertion and matching
func TestRadixTree(t *testing.T) {
	tree := NewRadixTree("unit-test")

	type data struct {
		keys []uint64
		data uint64
	}

	datas := []data{
		{[]uint64{1, 2, 3, 4, 5, 6}, 1},
		{[]uint64{1, 2, 3, 4}, 1},
		{[]uint64{1, 2, 3, 4}, 2},
		{[]uint64{1, 2, 3}, 3},
		{[]uint64{1, 2, 3, 5, 6}, 4},
		{[]uint64{1, 2, 3, 5, 6}, 3},
		{[]uint64{1, 2, 3, 7, 8}, 5},
		{[]uint64{2, 3, 4, 5}, 6},
		{[]uint64{2, 3, 4, 6}, 7},
		{[]uint64{4, 3, 9, 8}, 8},
		{[]uint64{4, 6, 1, 2}, 9},
		{[]uint64{4, 4, 1, 8}, 10},
	}
	// Test empty tree query
	result := tree.MatchAll([]uint64{1}, 100)
	require.Len(t, result, 0)
	// Test insertion
	for _, d := range datas {
		tree.Insert(d.keys, d.data, 0)
	}
	// Verify insertion results through queries
	type matchdata struct {
		key    []uint64
		result map[uint64]int
	}
	expectedResults := []matchdata{
		{
			key: []uint64{1, 2, 3, 4, 5, 6},
			result: map[uint64]int{
				1: 6,
				2: 4,
				3: 3,
				4: 3,
				5: 3,
			},
		},
		{
			key: []uint64{1, 2, 3, 5},
			result: map[uint64]int{
				1: 3,
				2: 3,
				3: 4,
				4: 4,
				5: 3,
			},
		},
		{
			key: []uint64{1, 2, 3, 9},
			result: map[uint64]int{
				1: 3,
				2: 3,
				3: 3,
				4: 3,
				5: 3,
			},
		},
		{
			key: []uint64{2, 3, 4, 5},
			result: map[uint64]int{
				6: 4,
				7: 3,
			},
		},
	}
	for i, match := range expectedResults {
		ret := tree.MatchAll(match.key, 100)
		require.Lenf(t, ret, len(match.result), "case %d not expected", i)
		for k, v := range match.result {
			rv, ok := ret[k]
			require.Truef(t, ok, "case %d on key %d not expected", i, k)
			require.Equalf(t, v, rv, "case %d on key %d not expected", i, k)
		}
	}
}

// TestNewRadixTree_Concurrency tests concurrent insert and query operations
func TestNewRadixTree_Concurrency(t *testing.T) {
	tree := NewRadixTree("concurrency-test")
	wg := sync.WaitGroup{}

	// Test concurrent insert and query operations
	for i := 0; i < 10000; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			n := rand.Intn(8) + 1
			keys := []uint64{}
			for j := 0; j < n; j++ {
				keys = append(keys, uint64(rand.Intn(10)))
			}
			tree.Insert(keys, uint64(n), 0)
		}()
		go func() {
			defer wg.Done()
			n := rand.Intn(8) + 1
			keys := []uint64{}
			for j := 0; j < n; j++ {
				keys = append(keys, uint64(rand.Intn(10)))
			}
			tree.MatchAll(keys, 10)
		}()
	}
}

// TestRadixTree_GCByInsert tests garbage collection triggered by insert operations
func TestRadixTree_GCByInsert(t *testing.T) {
	tree := NewRadixTree("ut-gc")
	testInterval := 50 * time.Millisecond
	SetExpireDuration(2 * testInterval)
	defer func() {
		SetExpireDuration(DefaultExpireDuration)
	}()
	key := []uint64{1, 2, 3, 4}
	tree.Insert(key, 1, nil)
	result := tree.MatchAll(key, 100)
	require.Len(t, result, 1)
	require.Equal(t, result[1], 4)
	time.Sleep(3 * testInterval)
	result = tree.MatchAll(key, 100)
	require.Len(t, result, 0)
	tree.Insert(key, 1, nil)
	time.Sleep(testInterval)
	key2 := []uint64{1, 2, 3}
	tree.Insert(key2, 2, nil)
	time.Sleep(testInterval)
	// data=1 will expire
	// node=4 will expire
	result2 := tree.MatchAll(key, 100)
	require.Len(t, result2, 1)
	require.Equal(t, 3, result2[2])
}

// TestRadixTree_GCByInsert_Data tests data-level garbage collection scenarios
func TestRadixTree_GCByInsert_Data(t *testing.T) {
	tree := NewRadixTree("ut-gc-data")
	testInterval := 10 * time.Millisecond
	SetExpireDuration(3 * testInterval)
	defer func() {
		SetExpireDuration(DefaultExpireDuration)
	}()

	tree.Insert([]uint64{1, 2, 3}, 1, nil)
	time.Sleep(testInterval)
	tree.Insert([]uint64{1, 2, 3}, 2, nil)
	time.Sleep(2 * testInterval)
	// data = 1 expired, but data=2 not expired
	ret := tree.MatchAll([]uint64{1, 2, 3}, 10)
	require.Len(t, ret, 1)
	require.Equal(t, 3, ret[2])
}

// TestNewRadixTree_Traversal tests tree traversal functionality and level information
func TestNewRadixTree_Traversal(t *testing.T) {
	type Result struct {
		path []uint64
		data []int
	}
	results := []Result{}
	handle := func(_ *RadixNode, snode *StackRadixNode) bool {
		datas := []int{}
		for k := range snode.node.datas {
			datas = append(datas, int(k))
		}
		sort.Ints(datas)
		results = append(results, Result{
			path: snode.prefixPath,
			data: datas,
		})
		logger.Debugf("prefix %v data %v ,results %v", snode.prefixPath, datas, results)
		return true
	}
	testInterval := 50 * time.Millisecond
	SetExpireDuration(2 * testInterval)
	defer func() {
		SetExpireDuration(DefaultExpireDuration)
	}()
	tree := NewRadixTree("ut-traversal")
	// Build test data structure
	/*     root
	 *   12[1,2,3,4]          34[5,6]
	 *  3[1,2,3]  4[4]      5[5]  6[6]
	 * 4[2] 5[3]
	 */
	expected := []Result{
		{[]uint64{1, 2}, []int{1, 2, 3, 4}},
		{[]uint64{3, 4}, []int{5, 6}},
		{[]uint64{1, 2, 3}, []int{1, 2, 3}},
		{[]uint64{1, 2, 4}, []int{4}},
		{[]uint64{3, 4, 5}, []int{5}},
		{[]uint64{3, 4, 6}, []int{6}},
		{[]uint64{1, 2, 3, 4}, []int{2}},
		{[]uint64{1, 2, 3, 5}, []int{3}},
	}
	for _, data := range []struct {
		key  []uint64
		data uint64
	}{
		{
			[]uint64{1, 2, 3}, 1,
		},
		{
			[]uint64{1, 2, 3, 4}, 2,
		},
		{
			[]uint64{1, 2, 3, 5}, 3,
		},
		{
			[]uint64{1, 2, 4}, 4,
		},
		{
			[]uint64{3, 4, 5}, 5,
		},
		{
			[]uint64{3, 4, 6}, 6,
		},
	} {
		tree.Insert(data.key, data.data, nil)
	}
	tree.Traversal(handle)
	require.Equal(t, len(expected), len(results))
	for _, res := range results {
		found := false
		for _, exp := range expected {
			if sliceEqual(res.path, exp.path) {
				found = true
				require.Truef(t, sliceEqual(res.data, exp.data),
					"result path %d data not matched, res data %v, expected data %v",
					res.path, res.data, exp.data)
				break
			}
		}
		require.Truef(t, found, "result path %d not found", res.path)
	}
	// Check tree info statistics
	require.Len(t, tree.info, 3)
	/*     root
	 *   12[1,2,3,4]          34[5,6]
	 *  3[1,2,3]  4[4]      5[5]  6[6]
	 * 4[2] 5[3]
	 */
	expectedInfo := map[int]*RadixTreeLevelInfo{
		1: {
			NodesCount:   2,
			PrefixTotal:  4,
			MaxPrefixLen: 2,
			MinPrefixLen: 2,
		},
		2: {
			NodesCount:   4,
			PrefixTotal:  4,
			MaxPrefixLen: 1,
			MinPrefixLen: 1,
		},
		3: {
			NodesCount:   2,
			PrefixTotal:  2,
			MaxPrefixLen: 1,
			MinPrefixLen: 1,
		},
	}
	for lv, info := range expectedInfo {
		resInfo, ok := tree.info[lv]
		require.Truef(t, ok, "level %d not expected", lv)
		require.Equalf(t, info, resInfo, "level %d not expected", lv)
	}
	// Test Clean Expired
	time.Sleep(2 * testInterval)
	tree.CleanExpireNode()
	// Clean result
	results = results[:0]
	tree.Traversal(handle)
	require.Len(t, results, 0)
	require.Len(t, tree.info, 0)
}

// TestRadixTree_Traversal_Ignore tests that insert and query operations are ignored during traversal
func TestRadixTree_Traversal_Ignore(t *testing.T) {
	tree := NewRadixTree("traversal-ignore")
	keys := []uint64{1, 2, 3}
	tree.Insert(keys, 1, nil)
	tBlock := make(chan struct{})
	handle := func(_ *RadixNode, _ *StackRadixNode) bool {
		<-tBlock
		return true
	}
	// Start async traversal, expected to block
	go func() {
		tree.Traversal(handle)
	}()
	waitTraversal := func(expected bool) bool {
		ready := false
		for i := 0; i < 10; i++ {
			if tree.isTraversal.Load() == expected {
				ready = true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		return ready
	}
	startTraversal := waitTraversal(true)
	require.True(t, startTraversal)
	// During traversal, insert and query operations should be ignored
	key2 := []uint64{3, 4, 5}
	tree.Insert(key2, 2, nil)
	result := tree.MatchAll(keys, 100)
	require.Len(t, result, 0)
	// Close block to end traversal
	close(tBlock)
	traversalFinish := waitTraversal(false)
	require.True(t, traversalFinish)
	// Can query results after traversal
	result = tree.MatchAll(keys, 100)
	require.Len(t, result, 1)
	// Insert was also ignored
	result = tree.MatchAll(key2, 100)
	require.Len(t, result, 0)
	// Re-insert after traversal
	tree.Insert(key2, 2, nil)
	result = tree.MatchAll(key2, 100)
	require.Len(t, result, 1)
}

// TestRadixTree_MatchAll_TopK tests top-K matching functionality with various scenarios
func TestRadixTree_MatchAll_TopK(t *testing.T) {
	tree := NewRadixTree("topk-tree")
	// Build test data
	datas := []struct {
		key  []uint64
		data uint64
	}{
		{[]uint64{1}, 100},
		{[]uint64{1, 2, 3}, 101},
		{[]uint64{1, 2, 3}, 102},
		{[]uint64{1, 2, 3, 4}, 103},
		{[]uint64{1, 2, 3, 4}, 104},
	}
	for _, d := range datas {
		tree.Insert(d.key, d.data, nil)
	}
	// Test matching
	// With k=3, 103 and 104 are deterministic, but 101 and 102 may vary
	// In current implementation, earlier inserts are evicted first, but map iteration has randomness
	matches := map[uint64]int{}
	for i := 0; i < 100; i++ {
		ret := tree.MatchAll([]uint64{1, 2, 3, 4}, 3)
		require.Len(t, ret, 3)
		require.Equal(t, 4, ret[103])
		require.Equal(t, 4, ret[104])
		// Count occurrence of each data
		for k := range ret {
			v, ok := matches[k]
			if !ok {
				v = 0
			}
			v++
			matches[k] = v
		}
	}
	// Expected all possibilities
	require.Len(t, matches, 4)
	// k <= 0 scenario - expect full match
	SetDefaultQueryTopK(10)
	defer SetDefaultQueryTopK(defaultQueryTopK)
	ret := tree.MatchAll([]uint64{1, 2, 3, 4}, 0)
	require.Len(t, ret, 5)
}

// generateBenchData creates random test data for benchmarking
func generateBenchData(ktotal, ltotal int) ([][]uint64, []uint64) {
	// Generate random test data
	// Each insertion uses random keys + loads
	keys := [][]uint64{}
	loads := []uint64{}
	for i := 0; i < ktotal; i++ {
		ksize := 1 + rand.Intn(8) // Simulate 4K split into 8x512 hashes
		key := make([]uint64, ksize)
		for j := 0; j < ksize; j++ {
			key[j] = uint64(rand.Int63())
		}
		keys = append(keys, key)
	}
	for i := 0; i < ltotal; i++ {
		loads = append(loads, uint64(rand.Intn(10000)))
	}
	return keys, loads
}

// BenchmarkNewRadixTree_Query benchmarks query performance
func BenchmarkNewRadixTree_Query(b *testing.B) {
	// Reduce log level for benchmarking
	logger.SetLevel(logger.ErrorLevel)
	// Build a tree with test data
	keyTotal := 1000 * 10000
	loadTotal := 10000
	tree := NewRadixTree("expire-bench")
	keys, loads := generateBenchData(keyTotal, loadTotal)
	for _, key := range keys {
		l := loads[rand.Intn(loadTotal)]
		tree.Insert(key, l, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := keys[rand.Intn(keyTotal)]
		tree.MatchAll(k, 10)
	}
}

// BenchmarkNewRadixTree_Parallel benchmarks concurrent query and insert operations
func BenchmarkNewRadixTree_Parallel(b *testing.B) {
	// Reduce log level for benchmarking
	logger.SetLevel(logger.ErrorLevel)

	keyTotal := 1000 * 10000
	loadTotal := 10000
	keys, loads := generateBenchData(keyTotal, loadTotal)

	tree := NewRadixTree("benchmark-test")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := keys[rand.Intn(keyTotal)]
			l := loads[rand.Intn(loadTotal)]
			// Typical scenario: query first then insert
			tree.MatchAll(k, 10)
			tree.Insert(k, l, 0)
		}
	})
}

// BenchmarkRadixTree_CleanExpireNode benchmarks garbage collection performance
func BenchmarkRadixTree_CleanExpireNode(b *testing.B) {
	// Reduce log level for benchmarking
	logger.SetLevel(logger.ErrorLevel)
	newTree := func(b *testing.B) *RadixTree {
		b.StopTimer()
		keyTotal := 1000 * 10000
		loadTotal := 10000
		tree := NewRadixTree("expire-bench")
		keys, loads := generateBenchData(keyTotal, loadTotal)
		for _, key := range keys {
			l := loads[rand.Intn(loadTotal)]
			tree.Insert(key, l, 0)
		}
		b.StartTimer()
		return tree
	}

	for i := 0; i < b.N; i++ {
		tree := newTree(b)
		tree.CleanExpireNode()
	}

}
