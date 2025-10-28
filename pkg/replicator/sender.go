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

package replicator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aigw-project/metadata-center/pkg/servicediscovery"
	"github.com/aigw-project/metadata-center/pkg/servicediscovery/types"
	"github.com/aigw-project/metadata-center/pkg/utils/helper"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

const (
	EventTypeHeader                    = "Event-Type"
	EventTypeCtxKey                    = "eventType"
	TraceIdHeader                      = "TraceId"
	MetaDataCenterServiceDiscoveryHost = "META_DATA_CENTER_SVC_DISC_HOST"
	ReplicaEventPath                   = "/v1/replica/event"
	ReplicaEventTargetPort             = "REPLICA_CLIENT_TARGET_PORT"
	ReplicaClientDialTimeout           = "REPLICA_CLIENT_DIAL_TIMEOUT"
	ReplicaClientRequestTimeout        = "REPLICA_CLIENT_REQUEST_TIMEOUT"
	ReplicaClientMaxIdleConns          = "REPLICA_CLIENT_MAX_IDLE_CONNS"
	ReplicaClientMaxIdleConnTimeout    = "REPLICA_CLIENT_IDLE_CONN_TIMEOUT"
	ReplicaClientKeepAlivePeriod       = "REPLICA_CLIENT_KEEPALIVE_PERIOD"
	ReplicaDnsLookUpInterval           = "REPLICA_DNS_LOOKUP_INTERVAL"
)

// Replicator handles sending replication events to other nodes
type Replicator struct {
	client           *http.Client
	serviceDiscovery types.ServiceDiscovery
	port             int
}

// replicator is the singleton instance of the replication client
var replicator *Replicator

// createDefaultHTTPClient creates a configured HTTP client for replication
// Uses environment variables for timeout and connection pool configuration
func createDefaultHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   helper.GetDurationFromEnv(ReplicaClientDialTimeout, 500*time.Millisecond),
		KeepAlive: helper.GetDurationFromEnv(ReplicaClientKeepAlivePeriod, 10*time.Second),
	}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxIdleConns:        helper.GetIntFromEnv(ReplicaClientMaxIdleConns, 1024),
		MaxConnsPerHost:     helper.GetIntFromEnv(ReplicaClientMaxIdleConns, 1024),
		MaxIdleConnsPerHost: helper.GetIntFromEnv(ReplicaClientMaxIdleConns, 1024),
		IdleConnTimeout:     helper.GetDurationFromEnv(ReplicaClientMaxIdleConnTimeout, 5*time.Minute),
	}
	return &http.Client{
		Transport: transport,
		Timeout:   helper.GetDurationFromEnv(ReplicaClientRequestTimeout, 1*time.Second),
	}
}

// Replicate sends an event to all available replication targets
// Panics if replicator is not initialized
func Replicate(c context.Context, eventType string, payload any) {
	if replicator == nil {
		panic("Replicator is not initialized, please call replicator.Init() first")
	}
	replicator.replicate(c, eventType, payload)
}

// replicate sends replication events to all available hosts
// Marshals payload and initiates concurrent requests with retry logic
func (r *Replicator) replicate(c context.Context, eventType string, payload any) {
	hosts := r.serviceDiscovery.GetHosts()
	if len(hosts) == 0 {
		logger.Warnf("Replicator: No available hosts to replicate to for event '%s'", eventType)
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Replicator: marshal payload error: %v", err)
		return
	}

	traceID := helper.GetTraceIDFromCtx(c)
	logger.Debugf("Replicating event to hosts: %v", hosts)
	for _, host := range hosts {
		go r.sendRequestWithRetry(c, host, traceID, eventType, body)
	}
}

// sendRequestWithRetry sends a replication request with retry logic
// Handles panics and retries failed requests up to maxAttempts
func (r *Replicator) sendRequestWithRetry(ctx context.Context, targetHost, traceID, eventType string, body []byte) {
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("Replicator: Recovered from panic in replicate goroutine for Host %s. Panic: %v", targetHost, p)
		}
	}()

	url := fmt.Sprintf("http://%s:%d%s", targetHost, r.port, ReplicaEventPath)

	maxAttempts := 2 // 1 initial attempt + 1 retry
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			logger.Errorf("Replicator: Failed to create request for Host %s, aborting. Error: %v", targetHost, err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(TraceIdHeader, traceID)
		req.Header.Set(EventTypeHeader, eventType)

		resp, err := r.client.Do(req)
		if err != nil {
			lastErr = err
			logger.Warnf("Replicator: Failed to send event to %s (attempt %d/%d): %v. Retrying...", targetHost, attempt+1, maxAttempts, err)
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			logger.Warnf("Replicator: Failed to read error response body from %s (attempt %d/%d): %v", targetHost, attempt+1, maxAttempts, readErr)
		}

		if resp.StatusCode == http.StatusOK {
			logger.Debugf("Replicator: Successfully replicated event to %s (attempt %d/%d)", targetHost, attempt+1, maxAttempts)
			resp.Body.Close()
			return
		}

		lastErr = fmt.Errorf("server returned non-200 status: %d, body: %q", resp.StatusCode, string(respBody))
		// Non-200 status codes trigger retry
		logger.Warnf("Replicator: Replication failed (attempt %d/%d): %v. Retrying...", targetHost, attempt+1, maxAttempts, lastErr)
		resp.Body.Close()
	}

	logger.Errorf("Replicator: Failed to send event to %s after %d attempts. Last error: %v", targetHost, maxAttempts, lastErr)
}

// Init initializes the replicator singleton with service discovery
// Must be called before using Replicate function
func Init() {
	dnsConfig := servicediscovery.DNSConfig{
		Domain:         os.Getenv(MetaDataCenterServiceDiscoveryHost),
		LookupInterval: helper.GetDurationFromEnv(ReplicaDnsLookUpInterval, 5*time.Second),
		GetLocalHosts:  helper.GetLocalHosts,
	}

	sd, err := servicediscovery.NewDNSDiscovery(dnsConfig)
	if err != nil {
		logger.Fatalf("Failed to initialize service discovery: %v", err)
	}

	replicator = &Replicator{
		client:           createDefaultHTTPClient(),
		serviceDiscovery: sd,
		port:             helper.GetIntFromEnv(ReplicaEventTargetPort, 80),
	}

	logger.Infof("Replicator initialized successfully.")
}
