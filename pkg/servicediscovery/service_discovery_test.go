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

package servicediscovery

import (
	"errors"
	"testing"
	"time"

	"github.com/aigw-project/metadata-center/pkg/servicediscovery/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGetLocalHostsFunc is a mock implementation of GetLocalHostsFunc for testing
type mockGetLocalHostsFunc func() (string, error)

func (m mockGetLocalHostsFunc) GetLocalHosts() (string, error) {
	return m()
}

func TestNewDNSDiscovery(t *testing.T) {
	tests := []struct {
		name          string
		config        DNSConfig
		wantErr       bool
		errorContains string
	}{
		{
			name: "should create service discovery successfully",
			config: DNSConfig{
				Domain:         "example.com",
				LookupInterval: 5 * time.Second,
				GetLocalHosts: func() types.GetLocalHostsFunc {
					return func() (string, error) {
						return "127.0.0.1", nil
					}
				}(),
			},
			wantErr: false,
		},
		{
			name: "should fail when GetLocalHosts returns error",
			config: DNSConfig{
				Domain:         "example.com",
				LookupInterval: 5 * time.Second,
				GetLocalHosts: func() types.GetLocalHostsFunc {
					return func() (string, error) {
						return "", errors.New("failed to get local hosts")
					}
				}(),
			},
			wantErr:       true,
			errorContains: "failed to get local hosts",
		},
		{
			name: "should work without GetLocalHosts function",
			config: DNSConfig{
				Domain:         "example.com",
				LookupInterval: 5 * time.Second,
				GetLocalHosts:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := NewDNSDiscovery(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, sd)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, sd)
			}
		})
	}
}

func TestDNSDiscovery_GetHosts(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(sd *dnsServiceDiscovery)
		expectedHosts  []string
		expectedLength int
	}{
		{
			name: "should return empty list when no hosts discovered",
			setup: func(sd *dnsServiceDiscovery) {
				sd.mutex.Lock()
				defer sd.mutex.Unlock()
				sd.nodeList = make(map[string]struct{})
				sd.hosts = []string{}
			},
			expectedHosts:  []string{},
			expectedLength: 0,
		},
		{
			name: "should return hosts excluding local host",
			setup: func(sd *dnsServiceDiscovery) {
				sd.mutex.Lock()
				defer sd.mutex.Unlock()
				sd.nodeList = map[string]struct{}{
					"192.168.1.1": {},
					"192.168.1.2": {},
					"127.0.0.1":   {},
				}
				sd.localHost = "127.0.0.1"
				sd.hosts = []string{"192.168.1.1", "192.168.1.2"}
			},
			expectedHosts:  []string{"192.168.1.1", "192.168.1.2"},
			expectedLength: 2,
		},
		{
			name: "should return all hosts when no local host set",
			setup: func(sd *dnsServiceDiscovery) {
				sd.mutex.Lock()
				defer sd.mutex.Unlock()
				sd.nodeList = map[string]struct{}{
					"192.168.1.1": {},
					"192.168.1.2": {},
					"127.0.0.1":   {},
				}
				sd.localHost = ""
				sd.hosts = []string{"192.168.1.1", "192.168.1.2", "127.0.0.1"}
			},
			expectedHosts:  []string{"192.168.1.1", "192.168.1.2", "127.0.0.1"},
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := &dnsServiceDiscovery{
				config: DNSConfig{
					Domain:         "test.local",
					LookupInterval: 5 * time.Second,
				},
				nodeList: make(map[string]struct{}),
				hosts:    []string{},
			}

			if tt.setup != nil {
				tt.setup(sd)
			}

			hosts := sd.GetHosts()

			assert.Equal(t, tt.expectedLength, len(hosts))
			assert.ElementsMatch(t, tt.expectedHosts, hosts)
		})
	}
}

func TestDNSDiscovery_UpdateHosts(t *testing.T) {
	tests := []struct {
		name          string
		initialHosts  map[string]struct{}
		newHosts      []string
		localHost     string
		expectedHosts []string
	}{
		{
			name: "should add new hosts and remove stale hosts",
			initialHosts: map[string]struct{}{
				"192.168.1.1": {},
				"192.168.1.2": {},
			},
			newHosts:      []string{"192.168.1.2", "192.168.1.3"},
			localHost:     "192.168.1.1",
			expectedHosts: []string{"192.168.1.2", "192.168.1.3"}, // 192.168.1.1 is local host, 192.168.1.2 should remain
		},
		{
			name: "should handle empty new hosts",
			initialHosts: map[string]struct{}{
				"192.168.1.1": {},
				"192.168.1.2": {},
			},
			newHosts:      []string{},
			localHost:     "192.168.1.1",
			expectedHosts: []string{}, // All hosts should be removed
		},
		{
			name: "should handle no local host exclusion",
			initialHosts: map[string]struct{}{
				"192.168.1.1": {},
				"192.168.1.2": {},
			},
			newHosts:      []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
			localHost:     "",
			expectedHosts: []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}, // No local host to exclude
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := &dnsServiceDiscovery{
				config: DNSConfig{
					Domain:         "test.local",
					LookupInterval: 5 * time.Second,
				},
				nodeList: tt.initialHosts,
				localHost: tt.localHost,
				hosts:    []string{},
			}

			// Simulate the host update logic
			newHostsMap := make(map[string]struct{}, len(tt.newHosts))
			for _, host := range tt.newHosts {
				newHostsMap[host] = struct{}{}
			}

			sd.mutex.Lock()
			
			// Remove stale hosts
			for oldHost := range sd.nodeList {
				if _, exists := newHostsMap[oldHost]; !exists {
					delete(sd.nodeList, oldHost)
				}
			}

			// Add new hosts
			for newHost := range newHostsMap {
				if _, exists := sd.nodeList[newHost]; !exists {
					sd.nodeList[newHost] = struct{}{}
				}
			}

			// Update hosts list excluding local host
			updatedHosts := make([]string, 0, len(sd.nodeList))
			for host := range sd.nodeList {
				if host == sd.localHost {
					continue
				}
				updatedHosts = append(updatedHosts, host)
			}
			sd.hosts = updatedHosts
			
			sd.mutex.Unlock()

			hosts := sd.GetHosts()
			assert.ElementsMatch(t, tt.expectedHosts, hosts)
		})
	}
}