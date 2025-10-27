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
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

func newInferenceRequest(id, backend, modelName, ip, cluster string, length int32) *InferenceRequest {
	return &InferenceRequest{
		Cluster:      cluster,
		RequestId:    id,
		PromptLength: length,
		Ip:           ip,
	}
}

func newDeletionInferenceRequest(id string) *DeletionInferenceRequest {
	return &DeletionInferenceRequest{
		RequestId: id,
	}
}

func TestNewLoadStats(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	ls := NewLoadStats()
	tests := []struct {
		name            string
		req             *InferenceRequest
		deletionReq     *DeletionInferenceRequest
		add             bool
		totalEnginesLen int32
		curQueueLen     int32
		curPromptLen    int32
	}{
		{
			name:            "add new req",
			req:             newInferenceRequest("1", "sglang", "deepseekr1", "192.168.1.1", "test_domain", 512),
			add:             true,
			totalEnginesLen: 1,
			curQueueLen:     1,
			curPromptLen:    512,
		},
		{
			name:            "duplicate req id",
			req:             newInferenceRequest("1", "sglang", "deepseekr1", "192.168.1.1", "test_domain", 512),
			add:             true,
			totalEnginesLen: 1,
			curQueueLen:     1,
			curPromptLen:    512,
		},
		{
			name:            "add new req2",
			req:             newInferenceRequest("2", "sglang", "deepseekr1", "192.168.1.1", "test_domain", 512),
			add:             true,
			totalEnginesLen: 1,
			curQueueLen:     2,
			curPromptLen:    1024,
		},
		{
			name:            "delete req",
			req:             newInferenceRequest("1", "sglang", "deepseekr1", "192.168.1.1", "test_domain", 0),
			deletionReq:     newDeletionInferenceRequest("1"),
			add:             false,
			totalEnginesLen: 1,
			curQueueLen:     1,
			curPromptLen:    512,
		},
		{
			name:            "delete id not exists",
			req:             newInferenceRequest("noID", "sglang", "deepseekr1", "192.168.1.1", "test_domain", 512),
			deletionReq:     newDeletionInferenceRequest("noID"),
			add:             false,
			totalEnginesLen: 1,
			curQueueLen:     1,
			curPromptLen:    512,
		},
		{
			name:            "domain not exists",
			req:             newInferenceRequest("2", "sglang", "qwen", "192.168.1.1", "test_domain_not_exists", 512),
			deletionReq:     newDeletionInferenceRequest("2"),
			add:             false,
			totalEnginesLen: -1,
		},
		{
			name:            "engine not exists",
			req:             newInferenceRequest("2", "sglang", "deepseekr1", "192.168.1.2", "test_domain", 512),
			deletionReq:     newDeletionInferenceRequest("2"),
			add:             false,
			totalEnginesLen: 1,
			curQueueLen:     -1,
			curPromptLen:    0,
		},
		{
			name:            "add test_domain2",
			req:             newInferenceRequest("3", "sglang", "qwen", "192.168.1.1", "test_domain2", 512),
			add:             true,
			totalEnginesLen: 1,
			curQueueLen:     1,
			curPromptLen:    512,
		},
		{
			name:            "test_domain2 add new ip",
			req:             newInferenceRequest("4", "sglang", "qwen", "192.168.1.2", "test_domain2", 512),
			add:             true,
			totalEnginesLen: 2,
			curQueueLen:     1,
			curPromptLen:    512,
		},
	}
	for _, tc := range tests {
		if tc.add {
			ls.AddRequest(tc.req)
		} else {
			ls.DeletePromptLength(tc.deletionReq)
			ls.DeleteRequest(tc.deletionReq)
		}

		ms := ls.GetModelStats(tc.req.Cluster)
		if ms == nil {
			require.Equalf(t, int32(-1), tc.totalEnginesLen, "case %s model match failed", tc.name)
			continue
		}
		require.Equalf(t, tc.totalEnginesLen, ms.Size(), "case %s engines size not expected", tc.name)
		v, ok := ms.Engines.Load(tc.req.Ip)
		if !ok {
			require.Equalf(t, int32(-1), tc.curQueueLen, "case %s match engine failed", tc.name)
			continue
		}
		es := v.(*EngineStats)
		require.Equalf(t, tc.curQueueLen, es.GetQueuedReqNum(), "case[%s] engines length not expected", tc.name)
		require.Equalf(t, tc.curPromptLen, es.GetPromptLength(), "case[%s] engines prompt length not expected", tc.name)
	}
}

func TestCronClean(t *testing.T) {
	loadStats := NewLoadStats()
	interval := 10 * time.Millisecond
	ticker := time.NewTicker(2 * interval)
	SetRequestExpireDuration(5 * interval)
	defer func() {
		SetRequestExpireDuration(DefaultRequestExpireDuration)
		ticker.Stop()
	}()
	go cronClean(ticker, loadStats)

	testRequestID := "test-id"
	domain := "test_domain"
	req := &InferenceRequest{
		Cluster:      domain,
		RequestId:    testRequestID,
		PromptLength: 512,
		Ip:           "192.168.100.1",
		CreateTime:   time.Now(),
	}

	loadStats.AddRequest(req)

	ms := loadStats.GetModelStats(req.Cluster)
	require.Equal(t, int32(1), ms.Size())
	time.Sleep(13 * interval)
	ms = loadStats.GetModelStats(req.Cluster)
	require.Nil(t, ms)
}

func TestLoadStats_GC(t *testing.T) {
	loadStats := NewLoadStats()
	interval := 10 * time.Millisecond
	SetRequestExpireDuration(4 * interval)
	defer func() {
		SetRequestExpireDuration(DefaultRequestExpireDuration)
	}()
	domain := "test_domain"

	datas := []struct {
		ip                 string
		interval           time.Duration
		ms_expected_count  int32
		es_expecetd_count  int32
		es_expecetd_prompt int32
	}{

		{"192.168.100.1", interval, 1, 1, 512},
		{"192.168.100.1", interval, 1, 2, 1024},

		{"192.168.100.1", 3 * interval, 1, 1, 512},

		{"192.168.100.1", 5 * interval, 1, 0, 0},

		{"192.168.100.1", 2 * interval, 1, 1, 512},

		{"192.168.100.2", 3 * interval, 2, 1, 512},

		{"", 5 * interval, 1, 0, 0},
	}
	for idx, d := range datas {
		if d.ip != "" {
			req := &InferenceRequest{
				Cluster:      domain,
				RequestId:    fmt.Sprintf("%d", rand.Int()),
				Ip:           d.ip,
				PromptLength: 512,
				CreateTime:   time.Now(),
			}

			loadStats.AddRequest(req)
			time.Sleep(d.interval)
			loadStats.GC()

			ms := loadStats.GetModelStats(req.Cluster)
			if d.ms_expected_count == 0 {
				require.Nil(t, ms)
				continue
			}
			require.NotNilf(t, ms, "case %d ip %s", idx, d.ip)
			require.Equalf(t, d.ms_expected_count, ms.Size(), "case %d ip %s", idx, d.ip)
			v, ok := ms.Engines.Load(d.ip)
			require.Truef(t, ok, "case %d ip %s", idx, d.ip)
			es := v.(*EngineStats)
			require.Equalf(t, d.es_expecetd_count, es.GetQueuedReqNum(), "case %d ip %s", idx, d.ip)
			require.Equalf(t, d.es_expecetd_prompt, es.GetPromptLength(), "case %d ip %s", idx, d.ip)
		}
	}
}

var (
	benchMarkRequest []*InferenceRequest

	domainTypes = []string{"test1.com", "test2.com"}
)

func TestMain(m *testing.M) {
	const numKeys = 1000
	benchMarkRequest = make([]*InferenceRequest, numKeys)
	for i := 0; i < numKeys; i++ {
		domain := "test_domain_" + strconv.Itoa(i%100)
		benchMarkRequest[i] = &InferenceRequest{
			Cluster:      domain,
			Ip:           "192.168.1." + strconv.Itoa(i%256),
			PromptLength: 512,
			RequestId:    strconv.Itoa(rand.Int()),
		}
	}
	os.Exit(m.Run())
}

func BenchmarkLoadStats_GetModelStats(b *testing.B) {
	logger.SetLevel(logger.ErrorLevel)
	loadStats := NewLoadStats()

	for _, req := range benchMarkRequest {
		loadStats.AddRequest(req)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := benchMarkRequest[i]
			_ = loadStats.GetModelStats(req.Cluster)
			i++
			if i >= len(benchMarkRequest) {
				i = 0
			}
		}
	})
}

func BenchmarkLoadStats_AddRequest(b *testing.B) {
	logger.SetLevel(logger.ErrorLevel)
	loadStats := NewLoadStats()

	for _, req := range benchMarkRequest {
		loadStats.AddRequest(req)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		turn := 0
		for pb.Next() {

			if i%2 == 0 {
				req := benchMarkRequest[i]
				loadStats.AddRequest(req)
			} else {

				req := &InferenceRequest{
					Cluster:      domainTypes[i%2],
					Ip:           "10.0.0." + strconv.Itoa(i%256),
					PromptLength: 512,
					RequestId:    strconv.Itoa(rand.Int()),
				}
				loadStats.AddRequest(req)
			}
			i++
			if i >= len(benchMarkRequest) {
				i = 0
				turn++
			}
		}
	})
}

func BenchmarkLoadStats_ReadWrite(b *testing.B) {
	logger.SetLevel(logger.ErrorLevel)
	loadStats := NewLoadStats()

	for i, req := range benchMarkRequest {
		if i%2 == 0 {
			loadStats.AddRequest(req)
		}
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		turn := 0
		for pb.Next() {
			if i%10 < 7 {
				req := benchMarkRequest[i]
				_ = loadStats.GetModelStats(req.Cluster)
			} else {
				req := &InferenceRequest{
					Cluster:      domainTypes[i%2],
					Ip:           "10.0.0." + strconv.Itoa(i%256),
					PromptLength: 512,
					RequestId:    strconv.Itoa(rand.Int()),
				}
				loadStats.AddRequest(req)
			}
			i++
			if i >= len(benchMarkRequest) {
				i = 0
				turn++
			}
		}
	})
}

func TestDescPromptLength(t *testing.T) {
	ls := NewLoadStats()
	interval := 10 * time.Millisecond
	ticker := time.NewTicker(2 * interval)
	SetRequestExpireDuration(5 * interval)
	defer func() {
		SetRequestExpireDuration(DefaultRequestExpireDuration)
		ticker.Stop()
	}()
	go cronClean(ticker, ls)

	domain := "test_domain"
	req1 := &InferenceRequest{
		Cluster:      domain,
		RequestId:    "test-id-1",
		PromptLength: 512,
		Ip:           "192.168.100.1",
		CreateTime:   time.Now(),
	}

	ms := ls.GetModelStats(req1.Cluster)
	assert.Nil(t, ms)
	ls.decEnginePromptLength(req1)

	ls.AddRequest(req1)
	ls.AddRequest(req1)
	ms = ls.GetModelStats(req1.Cluster)
	require.Equal(t, int32(1), ms.Size())
	v, ok := ms.Engines.Load(req1.Ip)
	if !ok {
		t.Errorf("%s engines not found", req1.Ip)
	}
	es := v.(*EngineStats)
	require.Equalf(t, int32(1), es.GetQueuedReqNum(), "%s engines length not expected", req1.Ip)
	require.Equalf(t, int32(512), es.GetPromptLength(), "%s engines prompt length not expected", req1.Ip)

	req2 := &InferenceRequest{
		Cluster:      domain,
		RequestId:    "test-id-2",
		PromptLength: 512,
		Ip:           "192.168.100.2",
		CreateTime:   time.Now(),
	}
	ls.decEnginePromptLength(req2)
	_, ok = ms.Engines.Load(req2.Ip)
	if ok {
		t.Errorf("%s engines should not found", req1.Ip)
	}
}

func TestDeleteNonExistentRequest(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	ls := NewLoadStats()

	cluster := "test_domain"

	req := &DeletionInferenceRequest{
		RequestId: "non-existent-id",
	}

	ls.DeleteRequest(req)

	ms := ls.GetModelStats(cluster)
	assert.Nil(t, ms)

	time.Sleep(1100 * time.Millisecond)

	ms = ls.GetModelStats(cluster)
	assert.Nil(t, ms)
}

func TestDeleteNonExistentPromptLength(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	ls := NewLoadStats()

	cluster := "test_domain"

	req := &DeletionInferenceRequest{
		RequestId: "non-existent-id",
	}

	ls.DeletePromptLength(req)

	ms := ls.GetModelStats(cluster)
	assert.Nil(t, ms)

	time.Sleep(1100 * time.Millisecond)

	ms = ls.GetModelStats(cluster)
	assert.Nil(t, ms)
}

func TestDelayDeleteWithLateAdd(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	ls := NewLoadStats()

	req := &InferenceRequest{
		Cluster:      "test_domain",
		RequestId:    "late-add-test-id",
		PromptLength: 512,
		Ip:           "192.168.1.1",
	}

	deletionReq := &DeletionInferenceRequest{
		RequestId: req.RequestId,
	}

	ls.DeleteRequest(deletionReq)

	time.Sleep(500 * time.Millisecond)
	ls.AddRequest(req)

	ms := ls.GetModelStats(req.Cluster)
	require.NotNil(t, ms)
	v, ok := ms.Engines.Load(req.Ip)
	require.True(t, ok)
	es := v.(*EngineStats)
	require.Equal(t, int32(1), es.GetQueuedReqNum())
	require.Equal(t, int32(512), es.GetPromptLength())

	time.Sleep(600 * time.Millisecond)

	ms = ls.GetModelStats(req.Cluster)
	require.NotNil(t, ms)
	v, ok = ms.Engines.Load(req.Ip)
	require.True(t, ok)
	es = v.(*EngineStats)
	require.Equal(t, int32(0), es.GetQueuedReqNum())
	require.Equal(t, int32(0), es.GetPromptLength())
}

func TestConcurrencyUpdateLoadAndPromptLength(t *testing.T) {
	ls := NewLoadStats()

	var wg sync.WaitGroup
	requestCount := 50

	for i := 0; i < requestCount; i++ {
		reqID := fmt.Sprintf("req-%d", i)

		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			req := &InferenceRequest{
				RequestId:    id,
				Cluster:      "test_domain",
				PromptLength: 512,
				Ip:           "192.168.100.1",
				CreateTime:   time.Now(),
			}
			ls.AddRequest(req)
		}(reqID)

		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			req := &DeletionInferenceRequest{
				RequestId: id,
			}
			ls.DeletePromptLength(req)
		}(reqID)

		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			req := newDeletionInferenceRequest(id)
			ls.DeleteRequest(req)
		}(reqID)
	}

	wg.Wait()

	time.Sleep(1100 * time.Millisecond)

	ms := ls.GetModelStats("test_domain")
	if ms != nil {
		v, ok := ms.Engines.Load("192.168.100.1")
		if ok {
			es := v.(*EngineStats)
			assert.Equal(t, int32(0), es.GetQueuedReqNum(), "Queue length should be 0 after all requests are processed")
			assert.Equal(t, int32(0), es.GetPromptLength(), "Prompt length should be 0 after all requests are processed")
		} else {
			t.Log("Engine has been removed as expected when no requests remain")
		}
	} else {
		t.Log("Model stats has been removed as expected when no requests remain")
	}
}
