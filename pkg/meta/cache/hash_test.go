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

// TestHash_PromptToHash tests the hash function with different input scenarios
func TestHash_PromptToHash(t *testing.T) {
	h := NewHash(&HashConfig{
		ChunkLen: 10,
	})
	buf := []uint64{}
	p1 := []byte{}
	buf = h.PromptToHash(p1, buf)
	require.Len(t, buf, 0)
	p1 = []byte("12345")
	buf = h.PromptToHash(p1, buf)
	require.Len(t, buf, 1)
	p1 = []byte("12345678901")
	buf = h.PromptToHash(p1, buf)
	require.Len(t, buf, 2)
}
