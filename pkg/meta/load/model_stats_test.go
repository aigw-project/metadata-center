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

	"github.com/stretchr/testify/require"

	"github.com/aigw-project/metadata-center/pkg/utils/helper"
)

func TestNewModelStats(t *testing.T) {
	ms := NewModelStats("unit-test")
	ip := "192.168.100.1"
	ip2 := "192.168.100.2"
	_ = ms.LoadOrStore(ip)
	_ = ms.LoadOrStore(ip)
	_ = ms.LoadOrStore(ip2)
	require.Equal(t, int32(2), ms.Size())
	ms.Delete(ip)
	ms.Delete(ip)
	require.Equal(t, int32(1), ms.Size())
	_, ok := ms.Load(ip2)
	require.True(t, ok)
	ms.Delete(ip2)
	_, ok = ms.Load(ip2)
	require.False(t, ok)
}

func TestModelStats_ToJSON(t *testing.T) {
	ms := NewModelStats("unit-test-json")
	ip := "192.168.100.1"
	es := ms.LoadOrStore(ip)
	req := &InferenceRequest{
		Cluster:      "test_domain",
		RequestId:    "test",
		PromptLength: 512,
		Ip:           "192.168.100.1",
	}
	es.IncrementQueuedReqNumAndPromptLength(req, req.PromptLength)

	engines := ms.ToEngines()

	require.Len(t, engines, 1)
	require.Equal(t, ip, engines[0].Ip)
	require.Equal(t, int32(1), engines[0].QueuedReqNum)
	require.Equal(t, int32(512), engines[0].PromptLength)
}

func BenchmarkModelStats_ToJSON(b *testing.B) {
	ms := NewModelStats("bench-tojson")
	ipam := helper.NewIPAM("192.168.100.0/16")
	for i := 0; i < 1000; i++ {
		ip, _ := ipam.Alloc()
		es := ms.LoadOrStore(ip)
		req := &InferenceRequest{
			Cluster:      "test_domain",
			RequestId:    "test",
			PromptLength: 512,
			Ip:           ip,
		}
		es.IncrementQueuedReqNumAndPromptLength(req, req.PromptLength)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ms.ToEngines()
		}
	})
}
