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


package prom

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// grepText searches for lines containing the given key in a string
func grepText(key string, data string) []string {
	lines := strings.Split(data, "\n")
	matched := []string{}

	for _, line := range lines {
		if strings.Contains(line, key) {
			matched = append(matched, line)
		}
	}
	return matched
}

func TestPromMetricDelete(t *testing.T) {
	name := "test"
	ModelEngineCount.WithLabelValues(name).Set(0.0)

	Call := func() string {
		req := httptest.NewRequest("GET", "http:/127.0.0.1", nil)
		w := httptest.NewRecorder()
		promhttp.Handler().ServeHTTP(w, req)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		return string(body)
	}
	modelMetircs := `model_engine_count{model_name="test"}`
	m1 := grepText(modelMetircs, Call())
	require.Len(t, m1, 1)
	ModelEngineCount.DeleteLabelValues(name)

	m2 := grepText(modelMetircs, Call())
	require.Len(t, m2, 0)
}

// setupTestEnvironment creates isolated registry and router for testing
func setupTestEnvironment() (*prometheus.Registry, *gin.Engine) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(queuedNumGauge, promptLengthGauge)

	r := gin.New()
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	return registry, r
}

func TestDNSLookupHostsMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(DNSLookupHosts)

	r := gin.New()
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	DNSLookupHosts.WithLabelValues("test-domain", "10.0.0.1").Set(1)
	DNSLookupHosts.WithLabelValues("test-domain", "10.0.0.2").Set(1)
	DNSLookupHosts.WithLabelValues("test-domain", "10.0.0.3").Set(1)
	mf, _ := registry.Gather()
	initialCount := 0
	for _, metric := range mf {
		initialCount += len(metric.Metric)
	}

	t.Run("delete all host metrics for specific domain", func(t *testing.T) {
		DNSLookupHosts.DeletePartialMatch(prometheus.Labels{"domain": "test-domain"})

		mf, _ := registry.Gather()
		count := 0
		for _, metric := range mf {
			count += len(metric.Metric)
		}
		assert.Equal(t, initialCount-3, count)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/metrics", nil)
		r.ServeHTTP(w, req)

		assert.NotContains(t, w.Body.String(), `domain="test-domain"`)
	})

	t.Run("delete metrics for non-existent domain", func(t *testing.T) {
		mf, _ := registry.Gather()
		preCount := 0
		for _, metric := range mf {
			preCount += len(metric.Metric)
		}

		DNSLookupHosts.DeletePartialMatch(prometheus.Labels{"domain": "nonexistent-domain"})

		mf, _ = registry.Gather()
		count := 0
		for _, metric := range mf {
			count += len(metric.Metric)
		}
		assert.Equal(t, preCount, count)
	})
}

// getMetricCount returns the total number of metrics in the registry
func getMetricCount(registry *prometheus.Registry) int {
	mf, _ := registry.Gather()
	count := 0
	for _, metric := range mf {
		count += len(metric.Metric)
	}
	return count
}

func TestDeleteEngineMetric(t *testing.T) {
	registry, r := setupTestEnvironment()

	queuedNumGauge.WithLabelValues("model1", "1.1.1.1").Set(10)
	promptLengthGauge.WithLabelValues("model1", "1.1.1.1").Set(20)
	queuedNumGauge.WithLabelValues("model1", "2.2.2.2").Set(30)
	initialCount := getMetricCount(registry)

	t.Run("delete existing engine metrics", func(t *testing.T) {
		DeleteEngineMetric("model1", "1.1.1.1")

		assert.Equal(t, initialCount-2, getMetricCount(registry))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/metrics", nil)
		r.ServeHTTP(w, req)

		assert.NotContains(t, w.Body.String(), `model_name="model1", engine_ip="1.1.1.1"`)
		assert.Contains(t, w.Body.String(), `engine_ip="2.2.2.2",model_name="model1"`)
	})

	t.Run("delete non-existent engine metrics", func(t *testing.T) {
		preCount := getMetricCount(registry)
		DeleteEngineMetric("invalid", "9.9.9.9")
		assert.Equal(t, preCount, getMetricCount(registry))
	})
}

func TestDeleteModelMetric(t *testing.T) {
	registry, r := setupTestEnvironment()

	queuedNumGauge.WithLabelValues("modelA", "1.1.1.1").Set(1)
	promptLengthGauge.WithLabelValues("modelA", "1.1.1.1").Set(2)
	queuedNumGauge.WithLabelValues("modelA", "2.2.2.2").Set(3)
	promptLengthGauge.WithLabelValues("modelB", "3.3.3.3").Set(4)
	initialCount := getMetricCount(registry)

	t.Run("delete existing model metrics", func(t *testing.T) {
		DeleteModelMetric("modelA")

		assert.Equal(t, initialCount-3, getMetricCount(registry))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/metrics", nil)
		r.ServeHTTP(w, req)

		t.Logf("body: %s", w.Body.String())
		assert.NotContains(t, w.Body.String(), "model_name=\"modelA\"")
		assert.Contains(t, w.Body.String(), "model_name=\"modelB\"")
	})
}

func TestSetReplicationLatencyMetric(t *testing.T) {
	registry, _ := setupTestEnvironment()

	registry.MustRegister(replicationLatency)
	t.Run("positive latency", func(t *testing.T) {
		nowNano := time.Now().UnixNano()
		pastNano := nowNano - int64(50*time.Millisecond)

		mf, _ := registry.Gather()
		var beforeCount int
		for _, metric := range mf {
			if *metric.Name == "replication_latency_ms" {
				if len(metric.Metric) > 0 && metric.Metric[0].Histogram != nil && metric.Metric[0].Histogram.SampleCount != nil {
					beforeCount = int(*metric.Metric[0].Histogram.SampleCount)
				}
				break
			}
		}

		SetReplicationLatencyMillisecond(pastNano, "test-request-id")

		mf, err := registry.Gather()
		assert.NoError(t, err)

		var found bool
		for _, metric := range mf {
			if *metric.Name == "replication_latency_ms" {
				if len(metric.Metric) > 0 && metric.Metric[0].Histogram != nil && metric.Metric[0].Histogram.SampleCount != nil {
					afterCount := int(*metric.Metric[0].Histogram.SampleCount)
					assert.GreaterOrEqual(t, afterCount, beforeCount)
				}
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("negative latency", func(t *testing.T) {
		nowNano := time.Now().UnixNano()
		futureNano := nowNano + int64(50*time.Millisecond)

		mf, _ := registry.Gather()
		var beforeCount int
		for _, metric := range mf {
			if *metric.Name == "replication_latency_ms" {
				if len(metric.Metric) > 0 && metric.Metric[0].Histogram != nil && metric.Metric[0].Histogram.SampleCount != nil {
					beforeCount = int(*metric.Metric[0].Histogram.SampleCount)
				}
				break
			}
		}

		SetReplicationLatencyMillisecond(futureNano, "test-request-id-negative")

		mf, err := registry.Gather()
		assert.NoError(t, err)

		var found bool
		for _, metric := range mf {
			if *metric.Name == "replication_latency_ms" {
				if len(metric.Metric) > 0 && metric.Metric[0].Histogram != nil && metric.Metric[0].Histogram.SampleCount != nil {
					afterCount := int(*metric.Metric[0].Histogram.SampleCount)
					assert.Equal(t, afterCount, beforeCount)
				}
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("invalid timestamp", func(t *testing.T) {
		// Use 0 or negative timestamp
		SetReplicationLatencyMillisecond(0, "test-request-id-invalid")
		SetReplicationLatencyMillisecond(-1, "test-request-id-invalid")

		assert.True(t, true)
	})
}
