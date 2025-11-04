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
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockServiceDiscovery is a mock implementation of types.ServiceDiscovery for testing
type mockServiceDiscovery struct {
	hosts []string
}

func (m *mockServiceDiscovery) GetHosts() []string {
	return m.hosts
}

func (m *mockServiceDiscovery) Start() error {
	return nil
}

func (m *mockServiceDiscovery) Stop() error {
	return nil
}

func TestReplicate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		cleanup     func()
		shouldPanic bool
	}{
		{
			name: "should panic when replicator is not initialized",
			setup: func() {
				replicator = nil
			},
			shouldPanic: true,
		},
		{
			name: "should not panic when replicator is initialized",
			setup: func() {
				mockSD := &mockServiceDiscovery{hosts: []string{"localhost"}}
				replicator = &Replicator{
					client:           &http.Client{},
					serviceDiscovery: mockSD,
					port:             8080,
				}
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("Expected panic but none occurred")
				}
			}()

			Replicate(context.Background(), "test.event", map[string]string{"key": "value"})
		})
	}
}

func TestReplicator_replicate(t *testing.T) {
	tests := []struct {
		name          string
		hosts         []string
		payload       interface{}
		shouldMarshal bool
	}{
		{
			name:          "should not send requests when no hosts available",
			hosts:         []string{},
			payload:       map[string]string{"key": "value"},
			shouldMarshal: true,
		},
		{
			name:          "should handle JSON marshaling error",
			hosts:         []string{"localhost"},
			payload:       func() {},
			shouldMarshal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSD := &mockServiceDiscovery{hosts: tt.hosts}
			replicator := &Replicator{
				client:           &http.Client{},
				serviceDiscovery: mockSD,
				port:             8080,
			}

			// This should not panic and should handle the cases appropriately
			replicator.replicate(context.Background(), "test.event", tt.payload)
		})
	}
}

func TestReplicator_sendRequestWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		ctx           context.Context
		shouldSucceed bool
	}{
		{
			name:          "should handle panic in goroutine",
			host:          "localhost",
			ctx:           context.Background(),
			shouldSucceed: false,
		},
		{
			name: "should handle request creation error",
			host: "localhost",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSD := &mockServiceDiscovery{hosts: []string{tt.host}}
			replicator := &Replicator{
				client:           &http.Client{},
				serviceDiscovery: mockSD,
				port:             8080,
			}

			// This should not panic and should handle the cases appropriately
			replicator.sendRequestWithRetry(tt.ctx, tt.host, "test-trace", "test.event", []byte(`{}`))
		})
	}
}

func TestCreateDefaultHTTPClient(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cleanup  func()
		validate func(t *testing.T, client *http.Client)
	}{
		{
			name: "should create client with default values",
			setup: func() {
				os.Unsetenv(ReplicaClientDialTimeout)
				os.Unsetenv(ReplicaClientKeepAlivePeriod)
				os.Unsetenv(ReplicaClientMaxIdleConns)
				os.Unsetenv(ReplicaClientMaxIdleConnTimeout)
				os.Unsetenv(ReplicaClientRequestTimeout)
			},
			validate: func(t *testing.T, client *http.Client) {
				assert.NotNil(t, client)
				assert.Equal(t, 1*time.Second, client.Timeout)

				transport := client.Transport.(*http.Transport)
				assert.Equal(t, 1024, transport.MaxIdleConns)
				assert.Equal(t, 1024, transport.MaxConnsPerHost)
				assert.Equal(t, 1024, transport.MaxIdleConnsPerHost)
				assert.Equal(t, 5*time.Minute, transport.IdleConnTimeout)
			},
		},
		{
			name: "should use environment variable values",
			setup: func() {
				os.Setenv(ReplicaClientDialTimeout, "2s")
				os.Setenv(ReplicaClientKeepAlivePeriod, "30s")
				os.Setenv(ReplicaClientMaxIdleConns, "512")
				os.Setenv(ReplicaClientMaxIdleConnTimeout, "10m")
				os.Setenv(ReplicaClientRequestTimeout, "5s")
			},
			cleanup: func() {
				os.Unsetenv(ReplicaClientDialTimeout)
				os.Unsetenv(ReplicaClientKeepAlivePeriod)
				os.Unsetenv(ReplicaClientMaxIdleConns)
				os.Unsetenv(ReplicaClientMaxIdleConnTimeout)
				os.Unsetenv(ReplicaClientRequestTimeout)
			},
			validate: func(t *testing.T, client *http.Client) {
				assert.NotNil(t, client)
				assert.Equal(t, 5*time.Second, client.Timeout)

				transport := client.Transport.(*http.Transport)
				assert.Equal(t, 512, transport.MaxIdleConns)
				assert.Equal(t, 512, transport.MaxConnsPerHost)
				assert.Equal(t, 512, transport.MaxIdleConnsPerHost)
				assert.Equal(t, 10*time.Minute, transport.IdleConnTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			client := createDefaultHTTPClient()
			tt.validate(t, client)
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cleanup  func()
		validate func(t *testing.T)
	}{
		{
			name: "should initialize replicator successfully",
			setup: func() {
				os.Setenv(MetaDataCenterServiceDiscoveryHost, "test.local")
				os.Setenv(ReplicaEventTargetPort, "8080")
				os.Setenv("POD_IP", "127.0.0.1") // Required for service discovery
			},
			cleanup: func() {
				os.Unsetenv(MetaDataCenterServiceDiscoveryHost)
				os.Unsetenv(ReplicaEventTargetPort)
				os.Unsetenv("POD_IP")
			},
			validate: func(t *testing.T) {
				assert.NotNil(t, replicator)
				assert.NotNil(t, replicator.client)
				assert.NotNil(t, replicator.serviceDiscovery)
				assert.Equal(t, 8080, replicator.port)
			},
		},
		{
			name: "should use default port when environment variable not set",
			setup: func() {
				os.Setenv(MetaDataCenterServiceDiscoveryHost, "test.local")
				os.Setenv("POD_IP", "127.0.0.1") // Required for service discovery
				os.Unsetenv(ReplicaEventTargetPort)
			},
			cleanup: func() {
				os.Unsetenv(MetaDataCenterServiceDiscoveryHost)
				os.Unsetenv("POD_IP")
			},
			validate: func(t *testing.T) {
				assert.NotNil(t, replicator)
				assert.Equal(t, 80, replicator.port)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton before each test
			replicator = nil

			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// This should not panic
			Init()

			tt.validate(t)
		})
	}
}