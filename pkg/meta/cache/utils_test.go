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

	"github.com/stretchr/testify/require"
)

// TestLongestCommonPrefix tests the longest common prefix calculation
func TestLongestCommonPrefix(t *testing.T) {
	a := []uint64{1, 2, 3}
	b := []uint64{1, 2}
	l := longestCommonPrefix(a, b)
	require.Equal(t, 2, l)
}

// TestSliceEqual tests slice equality comparison with various cases
func TestSliceEqual(t *testing.T) {
	a := []uint64{1, 2}
	a1 := []uint64{2, 3}
	a2 := []uint64{1}

	require.True(t, sliceEqual(a, a))
	require.False(t, sliceEqual(a, a1))
	require.False(t, sliceEqual(a, a2))
}

// TestIPConvert tests IP address conversion between string and integer formats
func TestIPConvert(t *testing.T) {
	ip := "192.168.100.1"
	v := IPStr2Int(ip)
	ip2 := IntToIP(v)
	require.Equal(t, ip, ip2)
	ip_invalid := ""
	v0 := IPStr2Int(ip_invalid)
	require.Equal(t, uint32(0), v0)
	p0 := IntToIP(0)
	require.Equal(t, "", p0)
}

// TestNewTopKMap tests TopKMap functionality with heap replacement logic
func TestNewTopKMap(t *testing.T) {
	for i := 0; i < 1000; i++ {
		m := NewTopKMap(3)
		m.Add(1, 1)
		m.Add(2, 1)
		m.Add(3, 1)
		// Trigger heap adjustment
		m.Add(1, 10)
		// Replace key 2 (first inserted in min-heap array implementation)
		m.Add(4, 10)
		ret := m.GetMap()
		require.Len(t, ret, 3)
		require.Equal(t, 10, ret[1])
		require.Equal(t, 1, ret[3])
		require.Equal(t, 10, ret[4])
	}
}
