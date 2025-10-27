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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

var (
	// ModelEngineCount tracks the number of engines for each model
	ModelEngineCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "model_engine_count",
			Help: "Number of engines for each model",
		},
		[]string{"model_name"},
	)
	// HttpRequestStatusCodeCount counts HTTP requests by status code
	HttpRequestStatusCodeCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_status_code_total",
			Help: "Count of HTTP requests by prom code",
		},
		[]string{"url", "method", "status_code"},
	)

	// HttpRequestDuration tracks HTTP request duration distribution histogram
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_us",
			Help:    "Histogram of HTTP request durations in microseconds",
			Buckets: prometheus.ExponentialBuckets(500, 2, 3),
		},
		[]string{"method", "url"}, // Labels: method and url
	)

	// queuedNumGauge tracks the queued request count per model and engine
	queuedNumGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queued_num",
			Help: "The queuedNum value for each model and engine combination",
		},
		[]string{"model_name", "engine_ip"},
	)

	// promptLengthGauge tracks the prompt length per model and engine
	promptLengthGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prompt_length",
			Help: "The prompt length value for each model and engine combination",
		},
		[]string{"model_name", "engine_ip"},
	)

	// AppVersionInfo provides application version information
	AppVersionInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_version_info",
			Help: "Version information about the application",
		},
		[]string{"version"},
	)

	// replicationLatency tracks data replication event latency distribution
	replicationLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "replication_latency_ms",
			Help:    "Histogram of data replication event latency in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 8),
		},
	)

	// DomainRequestsTotal counts total requests partitioned by domain
	DomainRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "domain_requests_total",
			Help: "Total number of requests received, partitioned by domain",
		},
		[]string{"domain"},
	)

	// MetacenterNodeAlive indicates if the metacenter node is alive
	MetacenterNodeAlive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "metacenter_node_alive",
		Help: "Indicates if the metacenter node is alive",
	})

	// DNSLookupHosts tracks hosts resolved by DNS lookup for domains
	DNSLookupHosts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_lookup_hosts",
			Help: "Hosts resolved by DNS lookup for domain",
		},
		[]string{"domain", "host"},
	)
)

// SetLoadMetric sets load metrics for a specific model and engine
func SetLoadMetric(name, ip string, queuedNum, Length int32) {
	queuedNumGauge.WithLabelValues(name, ip).Set(float64(queuedNum))
	promptLengthGauge.WithLabelValues(name, ip).Set(float64(Length))
}

// DeleteEngineMetric removes metrics for a specific engine
func DeleteEngineMetric(name, ip string) {
	queuedNumGauge.DeleteLabelValues(name, ip)
	promptLengthGauge.DeleteLabelValues(name, ip)
}

// DeleteModelMetric removes all metrics for a specific model
func DeleteModelMetric(name string) {
	label := prometheus.Labels{
		"model_name": name,
	}
	queuedNumGauge.DeletePartialMatch(label)
	promptLengthGauge.DeletePartialMatch(label)
}

// SetReplicationLatencyMillisecond records replication latency metrics
// Handles clock skew by ignoring negative latencies
func SetReplicationLatencyMillisecond(timestampNano int64, requestID string) {
	if timestampNano > 0 {
		nowNano := time.Now().UnixNano()
		latencyMillisecond := float64(nowNano-timestampNano) / 1_000_000.0

		if latencyMillisecond >= 0 {
			replicationLatency.Observe(latencyMillisecond)
			logger.Debugf("requestID=[%s] replication latency: %.2f ms, timestampNano: %d ns, currentNano: %d ns",
				requestID, latencyMillisecond, timestampNano, nowNano)
			return
		}

		// Maybe negative replication latency occurs when replica clock is ahead of source clock due to clock skew
		// Check logs for requestID to verify clock sync issues between servers
		logger.Debugf("requestID=[%s] Negative replication latency detected: %.2f ms, timestampNano: %d ns, currentNano: %d ns, not included in metrics",
			requestID, latencyMillisecond, timestampNano, nowNano)
	}
}
