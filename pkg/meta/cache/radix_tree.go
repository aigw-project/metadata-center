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
	"maps"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

type RadixNode struct {
	/* Key: hash uint64 for next hash
	 * Value: *RadixNode
	 */
	children map[uint64]*RadixNode
	/* Prefix represents the current stored prefix */
	prefix []uint64
	/* data stores data information
	 * Expected key is also uint64, used to represent location (IP32+IP32)
	 * value is the corresponding timestamp for expiration
	 */
	datas map[uint64]int64
	/* unix timestamps */
	updateTimestamp int64
	/* lock for node-level operations */
	// lock sync.Mutex
}

func NewRadixNode(prefix []uint64) *RadixNode {
	return &RadixNode{
		children:        map[uint64]*RadixNode{},
		prefix:          slices.Clone(prefix),
		datas:           map[uint64]int64{},
		updateTimestamp: time.Now().UnixNano(),
		// lock:            sync.Mutex{},
	}
}

// RadixTreeLevelInfo contains statistics for a specific level in the radix tree
type RadixTreeLevelInfo struct {
	NodesCount   int `json:"nodes_count"`    // Number of nodes at this level
	PrefixTotal  int `json:"prefix_total"`   // Total prefix count at this level
	MaxPrefixLen int `json:"max_prefix_len"` // Maximum prefix length at this level
	MinPrefixLen int `json:"min_prefix_len"` // Minimum prefix length at this level
}

// RadixTreeInfo contains aggregated information about the radix tree
type RadixTreeInfo struct {
	Info       map[int]*RadixTreeLevelInfo `json:"info"`        // Level information by depth
	UpdateTime time.Time                   `json:"update_time"` // Last update timestamp
}

// RadixTree implements a radix tree for efficient prefix-based storage and retrieval
type RadixTree struct {
	name string
	root *RadixNode
	lock sync.Mutex
	// isTraversal indicates if the tree is currently being traversed
	isTraversal atomic.Bool
	// info contains tree statistics by level
	info           map[int]*RadixTreeLevelInfo
	infoUpdateTime time.Time
}

// NewRadixTree creates a new radix tree with the given name
func NewRadixTree(name string) *RadixTree {
	return &RadixTree{
		name:        name,
		root:        NewRadixNode(nil),
		isTraversal: atomic.Bool{},
	}
}

// Insert adds a data entry to the radix tree
// All matching nodes during tree traversal need to insert data
// data is the unique index, value is the corresponding value (currently not needed)
func (t *RadixTree) Insert(key []uint64, data uint64, _ any) {
	if t.isTraversal.Load() {
		logger.Infof("tree %s is in traversal. insert ignore", t.name)
		return
	}

	// Use tree-level lock instead of node-level lock
	t.lock.Lock()
	defer t.lock.Unlock()

	current := t.root
	remainKeys := key
	var parent *RadixNode

	for {
		// no need more matches
		if len(remainKeys) == 0 {
			break
		}
		parent = current
		next := remainKeys[0]
		// parent.lock.Lock()
		v, ok := parent.children[next]
		if !ok {
			// No match found, insert new data
			n := NewRadixNode(remainKeys)
			n.datas[data] = n.updateTimestamp
			parent.children[next] = n
			logger.Debugf("tree %s inserted a new node:%d:%v", t.name, next, remainKeys)
			// parent.lock.Unlock()
			break
		}
		// parent.lock.Unlock()
		current = v
		// current.lock.Lock()
		// Match successful, perform longest common prefix matching
		currentPrefix := current.prefix
		commonLen := longestCommonPrefix(remainKeys, currentPrefix)
		// 1. If completely matches the current node's prefix
		if commonLen == len(currentPrefix) {
			// Update current node's data
			current.updateTimestamp = time.Now().UnixNano()
			current.datas[data] = current.updateTimestamp
			// Get remaining keys for next iteration
			remainKeys = remainKeys[commonLen:]
			// current.lock.Unlock()
			continue
		}
		// Since we have a lock throughout, no need to check if prefix was modified
		//if !sliceEqual(currentPrefix, current.prefixPath) {
		//	continue
		//}
		// 2. Need to split the node
		logger.Debugf("tree %s key %v trigger node split, raw %v, to %v, %v ; new %v",
			t.name, remainKeys,
			current.prefix, current.prefix[:commonLen], current.prefix[commonLen:],
			remainKeys[commonLen:])
		// First split current node into new common node and remaining nodes
		// During node splitting, parent still points to current node
		// During splitting it should be:
		// 1) parent -> current; node1 -> [node2->next,nodenew]
		// 2) parent -> node1 -> [node2->next ,nodenew]
		newCommonNode := NewRadixNode(current.prefix[:commonLen])
		newCommonNode.datas = maps.Clone(current.datas)
		newCommonNode.datas[data] = newCommonNode.updateTimestamp
		// Replace current with next in place
		// Safe because we have tree-level lock
		newCommonNode.children[current.prefix[commonLen]] = current
		current.prefix = current.prefix[commonLen:]
		current.updateTimestamp = time.Now().UnixNano()
		// Process remaining keys and insert after new node
		remainKeys = remainKeys[commonLen:]
		if len(remainKeys) > 0 {
			leftNode := NewRadixNode(remainKeys)
			leftNode.datas[data] = leftNode.updateTimestamp
			newCommonNode.children[remainKeys[0]] = leftNode
		}
		// Current node's parent needs to update data
		// parent.lock.Lock()
		parent.children[newCommonNode.prefix[0]] = newCommonNode
		// parent.lock.Unlock()
		// Can only release current's lock after parent is updated
		// current.lock.Unlock()
		return
	}
	return
}

// MatchAll returns all data that matches the common prefix of the key
// Returns:
// {data, matched prefix length}
// Currently uses uint64(ip32+ip32) to represent data, may need adjustment if adding content later
func (t *RadixTree) MatchAll(key []uint64, topK int) map[uint64]int {
	results := NewTopKMap(topK)
	if t.isTraversal.Load() {
		logger.Infof("tree %s is in traversal. match ignore", t.name)
		return results.GetMap()
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	current := t.root
	remainKeys := key
	depth := 0
	var parent *RadixNode

	for {
		// no need more matches
		if len(remainKeys) == 0 {
			break
		}
		parent = current
		next := remainKeys[0]
		// parent.lock.Lock()
		v, ok := parent.children[next]
		if !ok {
			// parent.lock.Unlock()
			break
		}
		current = v
		// current.lock.Lock()
		// Match successful
		// Check if current node has expired, if so perform GC
		// Since we have node-level expiration, if current node expires, subsequent nodes are expected to also expire
		now := time.Now().UnixNano()
		if now >= current.updateTimestamp+expireDuration {
			delete(parent.children, current.prefix[0])
			// parent.lock.Unlock()
			// current.lock.Unlock()
			break
		}
		//parent.lock.Unlock()
		// Check data expiration
		for key, ts := range current.datas {
			if now > ts+expireDuration {
				delete(current.datas, key)
			}
		}
		// Record current data, update to latest depth
		commonLen := longestCommonPrefix(remainKeys, current.prefix)
		depth += commonLen
		for idx, _ := range current.datas {
			results.Add(idx, depth)
		}
		// Continue searching downward
		remainKeys = remainKeys[commonLen:]
		// current.lock.Unlock()
	}
	return results.GetMap()
}

type StackRadixNode struct {
	node       *RadixNode
	prefixPath []uint64
	level      int
}

// StackRadixNodeHandle processes corresponding node information during traversal
// If returns false, this node will no longer participate in subsequent traversal
// Usage: During global GC, if node has expired, skip traversing subsequent nodes
// Node processing in Handle doesn't require locking, default traversal has full tree lock
type StackRadixNodeHandle func(current *RadixNode, n *StackRadixNode) bool

// Traversal performs level-order traversal of the entire tree and calls handle for each node
func (t *RadixTree) Traversal(handle StackRadixNodeHandle) {
	// Need a global lock to avoid being split during traversal
	// Only need to lock root, entire tree will be locked, can ignore node-level locks
	t.lock.Lock()
	t.isTraversal.Store(true)
	// Update info once per traversal
	info := map[int]*RadixTreeLevelInfo{}
	defer func() {
		t.isTraversal.Store(false)
		logger.Infof("tree %s finished traversal state", t.name)
		t.lock.Unlock()
	}()
	logger.Infof("tree %s entered traversal state", t.name)

	stack := []*StackRadixNode{
		{t.root, []uint64{}, 1},
	}
	for len(stack) > 0 {
		// Record information
		// pop
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		current := top.node
		level := top.level
		// Skip root node, it doesn't contain information
		lvInfo, ok := info[level]
		if !ok {
			lvInfo = &RadixTreeLevelInfo{
				MinPrefixLen: math.MaxInt,
			}
		}
		lvInfo.NodesCount += len(current.children)
		// Traverse nodes level by level, skip first root node (no information)
		for _, node := range current.children {
			// Direct append case: if prefix has enough capacity, append may cause overwrite
			// For example: top.prefixPath = {1,2} cap=3
			// Then between different nodes: append(top.prefixPath, A) and append(top.prefixPath, B)
			// Won't trigger expansion, only causes B to overwrite A
			path := make([]uint64, len(top.prefixPath), len(top.prefixPath)+len(node.prefix))
			copy(path, top.prefixPath)
			path = append(path, node.prefix...)
			plen := len(node.prefix)
			if lvInfo.MaxPrefixLen < plen {
				lvInfo.MaxPrefixLen = plen
			}
			if lvInfo.MinPrefixLen > plen {
				lvInfo.MinPrefixLen = plen
			}
			lvInfo.PrefixTotal += plen
			snode := &StackRadixNode{
				node:       node,
				prefixPath: path,
				level:      level + 1,
			}
			if handle(current, snode) {
				// Only non-leaf nodes continue to be pushed to stack for traversal
				if len(node.children) != 0 {
					stack = append(stack, snode)
				}
			}

		}
		if lvInfo.NodesCount > 0 {
			info[level] = lvInfo
		}
	}

	t.info = info
	t.infoUpdateTime = time.Now()
}

// CleanExpireNode is a wrapper for traversal calls
func (t *RadixTree) CleanExpireNode() {
	handle := func(current *RadixNode, snode *StackRadixNode) bool {
		now := time.Now().UnixNano()
		if now >= snode.node.updateTimestamp+expireDuration {
			delete(current.children, snode.node.prefix[0])
			return false
		}
		for key, ts := range snode.node.datas {
			if now >= ts+expireDuration {
				delete(snode.node.datas, key)
			}
		}
		return true
	}
	t.Traversal(handle)
}

// GetTreeInfo gets information from the last traversal
func (t *RadixTree) GetTreeInfo() *RadixTreeInfo {
	t.lock.Lock()
	defer t.lock.Unlock()
	return &RadixTreeInfo{
		maps.Clone(t.info),
		t.infoUpdateTime,
	}
}
