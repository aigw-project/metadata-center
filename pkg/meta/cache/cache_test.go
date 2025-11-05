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
	"time"

	"github.com/stretchr/testify/require"
)

// TestCronClean tests the periodic garbage collection functionality
func TestCronClean(t *testing.T) {
	cp := NewCachePool()
	interval := 10 * time.Millisecond
	ticker := time.NewTicker(2 * interval)
	SetExpireDuration(5 * interval)
	defer func() {
		SetExpireDuration(DefaultExpireDuration)
		ticker.Stop()
	}()
	go cronClean(cp, ticker)
	testKey := "test"
	testPrompt := []byte("test")
	cp.Save(testKey, testPrompt, &Location{
		NodeIP:        "192.168.100.2",
		CoordinatorIP: "192.168.100.2",
	})
	keys := []string{}
	cp.pool.Range(func(key, value any) bool {
		ks := key.(string)
		keys = append(keys, ks)
		return true
	})
	require.Len(t, keys, 1)
	time.Sleep(10 * interval)
	keys = keys[:0]
	cp.pool.Range(func(key, value any) bool {
		ks := key.(string)
		keys = append(keys, ks)
		return true
	})
	// Expected to be GC'd
	require.Len(t, keys, 0)
}

// TestCachePool_GC tests garbage collection with data expiration scenarios
func TestCachePool_GC(t *testing.T) {
	cp := NewCachePool()
	interval := 10 * time.Millisecond
	SetExpireDuration(3 * interval)
	defer func() {
		SetExpireDuration(DefaultExpireDuration)
	}()
	testKey := "test"
	testPrompt := []byte("test")
	// Data expiration scenario
	cp.Save(testKey, testPrompt, &Location{
		NodeIP:        "192.168.100.2",
		CoordinatorIP: "192.168.100.2",
	})
	time.Sleep(interval)
	cp.Save(testKey, testPrompt, &Location{
		NodeIP:        "192.168.100.3",
		CoordinatorIP: "192.168.100.3",
	})
	time.Sleep(2 * interval)
	cp.GC()
	keys := []string{}
	cp.pool.Range(func(key, value any) bool {
		ks := key.(string)
		keys = append(keys, ks)
		return true
	})
	require.Len(t, keys, 1)
}

// TestCachePool_Save tests memory reuse scenarios
func TestCachePool_Save(t *testing.T) {
	cp := NewCachePool()
	llvm := "vllm_aigw-vllm-test"
	cp.Save(llvm, []byte("prompt1"), &Location{
		NodeIP: "10.238.8.26",
	})
	pd := "coordinator_pd-fc-test"
	cp.Save(pd, []byte("prompt2"), &Location{
		NodeIP: "10.238.8.26",
	})
	// Simulate concurrent retrieval of previous buf
	buf := hashPool.Get().([]uint64)
	defer func() {
		buf = buf[:0]
		hashPool.Put(buf)
	}()
	// Second memory reuse
	cp.Save(llvm, []byte("prompt1"), &Location{
		NodeIP: "10.238.8.26",
	})
}

// TestQuery tests cache query functionality with various scenarios
func TestQuery(t *testing.T) {
	type args struct {
		sp *SaveParam
		qp *QueryParam
	}
	tests := []struct {
		name      string
		args      args
		want      Location
		expectLen int
	}{
		{
			name: "raw prompt",
			args: args{
				sp: &SaveParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						Prompt: []byte("test"),
					},
					Location: &Location{
						NodeIP: "127.0.0.1",
					},
				},
				qp: &QueryParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						Prompt: []byte("test"),
					},
					TopK: 5,
				},
			},
			want: Location{
				NodeIP: "127.0.0.1",
			},
			expectLen: 1,
		},
		{
			name: "prompt hash",
			args: args{
				sp: &SaveParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072, 0xd81576c558f5841c, 0x4e3072cf8e5193bd},
					},
					Location: &Location{
						NodeIP: "127.0.0.1",
					},
				},
				qp: &QueryParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072, 0xd81576c558f5841c, 0x4e3072cf8e5193bd},
					},
					TopK: 5,
				},
			},
			want: Location{
				NodeIP: "127.0.0.1",
			},
			expectLen: 3,
		},
		{
			name: "prompt hash is prefix",
			args: args{
				sp: &SaveParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072, 0xd81576c558f5841c, 0x4e3072cf8e5193bd},
					},
					Location: &Location{
						NodeIP: "127.0.0.1",
					},
				},
				qp: &QueryParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072},
					},
					TopK: 5,
				},
			},
			want: Location{
				NodeIP: "127.0.0.1",
			},
			expectLen: 1,
		},
		{
			name: "prompt hash is longer",
			args: args{
				sp: &SaveParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072},
					},
					Location: &Location{
						NodeIP: "127.0.0.1",
					},
				},
				qp: &QueryParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072, 0xd81576c558f5841c, 0x4e3072cf8e5193bd},
					},
					TopK: 5,
				},
			},
			want: Location{
				NodeIP: "127.0.0.1",
			},
			expectLen: 1,
		},
		{
			name: "prompt hash not match",
			args: args{
				sp: &SaveParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072},
					},
					Location: &Location{
						NodeIP: "127.0.0.1",
					},
				},
				qp: &QueryParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{1111},
					},
					TopK: 5,
				},
			},
			want: Location{
				NodeIP: "",
			},
			expectLen: 0,
		},
		{
			name: "prompt hash not match",
			args: args{
				sp: &SaveParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{0x7646472539e54072},
					},
					Location: &Location{
						NodeIP: "127.0.0.1",
					},
				},
				qp: &QueryParam{
					Domain: "test.com",
					CommonCacheParams: CommonCacheParams{
						PromptHash: []uint64{1111},
					},
					TopK: 5,
				},
			},
			want: Location{
				NodeIP: "",
			},
			expectLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePool = NewCachePool()
			defer func() {
				cachePool = nil
			}()
			Save(tt.args.sp)
			got := Query(tt.args.qp)
			t.Logf("got: %v", got)
			for k, v := range got {
				l := Location{}
				l.Decode(k)
				t.Logf("location: %v", l)
				require.Equal(t, tt.want.NodeIP, l.NodeIP)
				if tt.args.sp.Prompt != nil {
					require.Equal(t, (len(tt.args.sp.Prompt)+511)/512, v)
				} else {
					require.Equal(t, tt.expectLen, v)
				}
			}
		})
	}
}
